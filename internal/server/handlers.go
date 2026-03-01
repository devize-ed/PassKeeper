package server

import (
	"context"
	"errors"
	"passKeper/internal/logger"
	"passKeper/internal/repository/db"
	"passKeper/internal/service"
	pb "passKeper/pkg/api"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Register is the gRPC handler for user registration.
func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
	logger.Log.Debugf("Registering a new user with username: %s", req.Username)
	// Call the CreateUser method from the AuthService.
	err := s.aService.CreateUser(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, db.ErrUserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	logger.Log.Debugf("User registered successfully with username: %s", req.Username)
	return &emptypb.Empty{}, nil
}

// Login is the gRPC handler for user login.
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	logger.Log.Debugf("Logging in a user with username: %s", req.Username)
	// Call the LoginUser method from the AuthService.
	token, err := s.aService.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to login user: %v", err)
	}

	logger.Log.Debugf("User logged in successfully with username: %s", req.Username)
	return &pb.LoginResponse{Token: token}, nil
}

// CreateItem is the gRPC handler for CreateItem requests.
func (s *ItemServer) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	logger.Log.Debugf("Creating a new item with type: %d", req.Type)
	// Marshal the item data to a byte array.
	data, err := proto.Marshal(req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal item data: %v", err)
	}
	// Call the CreateItem method from the ItemService.
	item, err := s.iService.CreateItem(ctx, int32(req.Type), data)
	if err != nil {
		if errors.Is(err, service.ErrInvalidItemType) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid item type: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create item: %v", err)
	}
	logger.Log.Debugf("Item created successfully with type: %d", req.Type)
	// Convert the item to a protobuf message.
	itemPb := &pb.Item{
		Id:        item.ID,
		Type:      pb.ItemType(item.Type),
		Data:      req.Data,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}
	return &pb.CreateItemResponse{Item: itemPb}, nil
}

// UpdateItem is the gRPC handler for UpdateItem requests.
func (s *ItemServer) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	logger.Log.Debugf("Updating an item with ID: %s and type: %d", req.Id, req.Type)
	// Marshal the item data to a byte array.
	data, err := proto.Marshal(req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal item data: %v", err)
	}
	// Truncate the timestamp to the microsecond
	timestamp := req.UpdatedAt.AsTime().UTC().Truncate(time.Microsecond)
	// Call the UpdateItem method from the ItemService.
	item, err := s.iService.UpdateItem(ctx, req.Id, int32(req.Type), data, timestamp)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
		}
		if errors.Is(err, db.ErrTimestampTooOld) {
			return nil, status.Errorf(codes.FailedPrecondition, "timestamp is older than the updated at: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update item: %v", err)
	}
	logger.Log.Debugf("Item updated successfully with ID: %s and type: %d", req.Id, req.Type)
	// Convert the item to a protobuf message.
	itemPb := &pb.Item{
		Id:        item.ID,
		Type:      pb.ItemType(item.Type),
		Data:      req.Data,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}
	return &pb.UpdateItemResponse{Item: itemPb}, nil
}

// DeleteItem is the gRPC handler for DeleteItem requests (soft delete).
func (s *ItemServer) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*emptypb.Empty, error) {
	logger.Log.Debugf("Deleting an item with ID: %s", req.Id)
	// Call the DeleteItem method from the ItemService.
	err := s.iService.DeleteItem(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete item: %v", err)
	}
	logger.Log.Debugf("Item deleted successfully with ID: %s", req.Id)
	return &emptypb.Empty{}, nil
}

// GetItem is the gRPC handler for GetItem requests.
func (s *ItemServer) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	logger.Log.Debugf("Getting an item with ID: %s", req.Id)
	// Call the GetItem method from the ItemService.
	item, err := s.iService.GetItem(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get item: %v", err)
	}
	logger.Log.Debugf("Item retrieved successfully with ID: %s", req.Id)
	// Unmarshal the item data.
	dataPb := &pb.ItemData{}
	err = proto.Unmarshal(item.Data, dataPb)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal item data: %v", err)
	}
	// Convert the item to a protobuf message.
	itemPb := &pb.Item{
		Id:        item.ID,
		Type:      pb.ItemType(item.Type),
		Data:      dataPb,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}
	return &pb.GetItemResponse{Item: itemPb}, nil
}

// ListItems is the gRPC handler for ListItems requests.
func (s *ItemServer) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	logger.Log.Debugf("Listing items with type: %d", req.Type)
	// Call the ListItems method from the ItemService.
	items, err := s.iService.GetAllItems(ctx, int32(req.Type))
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to list items: %v", err)
	}
	logger.Log.Debugf("Items listed successfully with type: %d", req.Type)
	// Convert the items to pointers.
	itemPtrs := make([]*pb.Item, len(items))
	// Unmarshal the item data.
	for i := range items {
		dataPb := &pb.ItemData{}
		if len(items[i].Data) > 0 {
			if err := proto.Unmarshal(items[i].Data, dataPb); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to unmarshal item data: %v", err)
			}
		}
		// Convert the item to a protobuf message.
		itemPtrs[i] = &pb.Item{
			Id:        items[i].ID,
			Type:      pb.ItemType(items[i].Type),
			Data:      dataPb,
			CreatedAt: timestamppb.New(items[i].CreatedAt),
			UpdatedAt: timestamppb.New(items[i].UpdatedAt),
		}
	}
	return &pb.ListItemsResponse{Items: itemPtrs}, nil
}
