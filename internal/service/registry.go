package service

import (
	"context"
	"fmt"

	"github.com/pentops/j5/builder"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/registry/internal/gen/j5/registry/registry/v1/registry_spb"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type ImageProvider interface {
	GetJ5Image(ctx context.Context, orgName, imageName, version string) (*source_j5pb.SourceImage, error)
}

type RegistryService struct {
	store ImageProvider

	registry_spb.UnimplementedDownloadServiceServer
}

func NewRegistryService(store ImageProvider) *RegistryService {
	return &RegistryService{
		store: store,
	}
}

func (s *RegistryService) RegisterGRPC(srv *grpc.Server) {
	registry_spb.RegisterDownloadServiceServer(srv, s)
}

func (s *RegistryService) DownloadImage(ctx context.Context, req *registry_spb.DownloadImageRequest) (*httpbody.HttpBody, error) {
	img, err := s.store.GetJ5Image(ctx, req.Owner, req.Name, req.Version)
	if err != nil {
		return nil, err
	}

	if img == nil {
		return nil, fmt.Errorf("image not found")
	}

	data, err := proto.Marshal(img)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: "application/octet-stream",
		Data:        data,
	}, nil
}

func (s *RegistryService) DownloadSwagger(ctx context.Context, req *registry_spb.DownloadSwaggerRequest) (*httpbody.HttpBody, error) {
	img, err := s.store.GetJ5Image(ctx, req.Owner, req.Name, req.Version)
	if err != nil {
		return nil, err
	}

	if img == nil {
		return nil, fmt.Errorf("image not found")
	}

	descriptorAPI, err := builder.DescriptorFromSource(img)
	if err != nil {
		return nil, err
	}

	asJson, err := builder.SwaggerFromDescriptor(descriptorAPI)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: "application/json",
		Data:        asJson,
	}, nil
}

func (s *RegistryService) DownloadJDef(ctx context.Context, req *registry_spb.DownloadJDefRequest) (*httpbody.HttpBody, error) {
	img, err := s.store.GetJ5Image(ctx, req.Owner, req.Name, req.Version)
	if err != nil {
		return nil, err
	}

	if img == nil {
		return nil, fmt.Errorf("image not found")
	}

	descriptorAPI, err := builder.DescriptorFromSource(img)
	if err != nil {
		return nil, err
	}

	jDefJSONBytes, err := builder.JDefFromDescriptor(descriptorAPI)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: "application/json",
		Data:        jDefJSONBytes,
	}, nil
}

func (s *RegistryService) DownloadClientAPI(ctx context.Context, req *registry_spb.DownloadClientAPIRequest) (*registry_spb.DownloadClientAPIResponse, error) {
	img, err := s.store.GetJ5Image(ctx, req.Owner, req.Name, req.Version)
	if err != nil {
		return nil, err
	}

	if img == nil {
		return nil, fmt.Errorf("image not found")
	}

	descriptorAPI, err := builder.DescriptorFromSource(img)
	if err != nil {
		return nil, err
	}

	return &registry_spb.DownloadClientAPIResponse{
		Api:     descriptorAPI,
		Version: img.GetVersion(),
	}, nil
}
