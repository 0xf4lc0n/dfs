package services

import (
	"context"
	"dfs/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type GrpcStorageClient struct {
	logger     *zap.Logger
	client     proto.StorageClient
	connection *grpc.ClientConn
}

func NewGrpcStorageClient(logger *zap.Logger) *GrpcStorageClient {
	return &GrpcStorageClient{logger: logger}
}

func (rsc *GrpcStorageClient) Connect(address string) error {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	rsc.connection = conn
	rsc.client = proto.NewStorageClient(conn)

	return nil
}

func (rsc *GrpcStorageClient) Disconnect() {
	if err := rsc.connection.Close(); err != nil {
		rsc.logger.Error("Cannot close connection to Grpc server", zap.Error(err))
	}
}

func (rsc *GrpcStorageClient) CreateHomeDirectory(dir *proto.HomeDir) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.CreateHomeDirectory(ctx, dir)

	if err != nil {
		rsc.logger.Error("Cannot create home directory", zap.Error(err))
		return false
	}

	return result.Success
}

func (rsc *GrpcStorageClient) GetOwnedFile(req *proto.OwnedFileRequest) *proto.FileEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.GetOwnedFile(ctx, req)

	if err != nil {
		rsc.logger.Error("Cannot get owned file", zap.Error(err))
		return nil
	}

	return result
}

func (rsc *GrpcStorageClient) GetFileById(req *proto.GetFileByIdRequest) *proto.FileEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.GetFileById(ctx, req)

	if err != nil {
		rsc.logger.Error("Cannot get file by id", zap.Error(err))
		return nil
	}

	return result
}

func (rsc *GrpcStorageClient) GetFileByUniqueName(file *proto.FileUniqueName) *proto.FileEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.GetFileByUniqueName(ctx, file)

	if err != nil {
		rsc.logger.Error("Cannot get file by id", zap.Error(err))
		return nil
	}

	return result
}

func (rsc *GrpcStorageClient) SaveFileOnDisk(req *proto.SaveFileRequest) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.SaveFileOnDisk(ctx, req)

	if err != nil {
		rsc.logger.Error("Cannot get file by id", zap.Error(err))
		return false
	}

	return result.Success
}

func (rsc *GrpcStorageClient) DeleteFileFromDisk(req *proto.DeleteFileRequest) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.DeleteFileFromDisk(ctx, req)

	if err != nil {
		rsc.logger.Error("Cannot get file by id", zap.Error(err))
		return false
	}

	return result.Success
}

func (rsc *GrpcStorageClient) GetFileContentFromDisk(req *proto.ReadFileRequest) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := rsc.client.GetFileContentFromDisk(ctx, req)

	if err != nil {
		rsc.logger.Error("Cannot get file by id", zap.Error(err))
		return nil
	}

	return result.Content
}

func (rsc *GrpcStorageClient) GetStoredFiles() *proto.StoredFiles {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storedFiles, err := rsc.client.GetStoredFiles(ctx, &emptypb.Empty{})

	if err != nil {
		rsc.logger.Error("Cannot get all stored files", zap.Error(err))
		return nil
	}

	return storedFiles
}

func (rsc *GrpcStorageClient) SyncStoredFiles(storedFiles *proto.StoredFiles) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := rsc.client.SyncStoredFiles(ctx, storedFiles)

	if err != nil {
		rsc.logger.Error("Cannot sync files", zap.Error(err))
		return false
	}

	return result.Success
}
