package services

import (
	"context"
	"dfs/proto"
	"dfs/storage/database"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRpcStorageServer struct {
	proto.UnimplementedStorageServer
	logger      *zap.Logger
	fileService *FileService
	storageRepo *database.StorageRepository
}

func NewGrpcStorageServer(log *zap.Logger, fileSrv *FileService, storageRepo *database.StorageRepository) *GRpcStorageServer {
	return &GRpcStorageServer{logger: log, fileService: fileSrv, storageRepo: storageRepo}
}

func (rss *GRpcStorageServer) CreateHomeDirectory(_ context.Context, homeDir *proto.HomeDir) (*proto.StorageResult, error) {
	result := rss.fileService.CreateDirectory(homeDir.Name)
	return &proto.StorageResult{Success: result}, nil
}

func (rss *GRpcStorageServer) GetOwnedFile(_ context.Context, req *proto.OwnedFileRequest) (*proto.FileEntry, error) {
	fileEntry := rss.storageRepo.GetOwnedFileById(req.FileId, uint(req.OwnerId))

	if fileEntry == nil {
		return nil, errors.New("cannot get owned file")
	}

	return &proto.FileEntry{Id: uint64(fileEntry.Id), OwnerId: uint64(fileEntry.OwnerId), Name: fileEntry.Name,
		UniqueName: fileEntry.UniqueName, CreationDate: timestamppb.New(fileEntry.CreationDate)}, nil
}

func (rss *GRpcStorageServer) GetFileById(_ context.Context, req *proto.GetFileByIdRequest) (*proto.FileEntry, error) {
	fileEntry := rss.storageRepo.GetFileById(req.FileId)

	if fileEntry == nil {
		return nil, errors.New("cannot get owned file")
	}

	return &proto.FileEntry{Id: uint64(fileEntry.Id), OwnerId: uint64(fileEntry.OwnerId), Name: fileEntry.Name,
		UniqueName: fileEntry.UniqueName, CreationDate: timestamppb.New(fileEntry.CreationDate)}, nil
}

func (rss *GRpcStorageServer) GetFileByUniqueName(_ context.Context, file *proto.FileUniqueName) (*proto.FileEntry, error) {
	fileEntry := rss.storageRepo.GetFileByUniqueName(file.Name)

	if fileEntry == nil {
		return nil, errors.New("cannot get file by unique name")
	}

	return &proto.FileEntry{Id: uint64(fileEntry.Id), OwnerId: uint64(fileEntry.OwnerId), Name: fileEntry.Name,
		UniqueName: fileEntry.UniqueName, CreationDate: timestamppb.New(fileEntry.CreationDate)}, nil
}

func (rss *GRpcStorageServer) SaveFileOnDisk(_ context.Context, req *proto.SaveFileRequest) (*proto.StorageResult, error) {
	saveResult := rss.fileService.EncryptAndSaveFile(req.SavePath, req.Content, req.EncryptionKey)
	return &proto.StorageResult{Success: saveResult}, nil
}

func (rss *GRpcStorageServer) DeleteFileFromDisk(_ context.Context, req *proto.DeleteFileRequest) (*proto.StorageResult, error) {
	deleteResult := rss.fileService.RemoveFileFromDisk(req.FilePath)
	return &proto.StorageResult{Success: deleteResult}, nil
}

func (rss *GRpcStorageServer) GetFileContentFromDisk(_ context.Context, req *proto.ReadFileRequest) (*proto.FileContent, error) {
	fileContent := rss.fileService.DecryptAndReadFileContent(req.ReadPath, req.DecryptionKey)

	if fileContent == nil {
		return nil, errors.New("cannot read file from disk")
	}

	return &proto.FileContent{Content: fileContent}, nil
}

func (rss *GRpcStorageServer) GetStoredFiles(context.Context, *emptypb.Empty) (*proto.StoredFiles, error) {
	storedFiles := &proto.StoredFiles{FilesPath: []string{}, FilesContent: [][]byte{}}

	storedFilesInfo := rss.fileService.GetStoredFiles()

	for p, c := range storedFilesInfo {
		storedFiles.FilesPath = append(storedFiles.FilesPath, p)
		storedFiles.FilesContent = append(storedFiles.FilesContent, c)
	}

	return storedFiles, nil
}

func (rss *GRpcStorageServer) SyncStoredFiles(_ context.Context, storedFiles *proto.StoredFiles) (*proto.StorageResult, error) {
	for i := 0; i < len(storedFiles.FilesPath); i++ {
		filePath := storedFiles.FilesPath[i]
		fileContent := storedFiles.FilesContent[i]

		rss.fileService.CreateMissingDirs(filePath)
		rss.fileService.SaveFileOnDisk(filePath, fileContent)
	}

	return &proto.StorageResult{Success: true}, nil
}
