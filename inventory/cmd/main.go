package main

import (
	"context"
	"errors"
	"fmt"
	inventoryV1 "github.com/alexander-kartavtsev/spaceship/shared/pkg/proto/inventory/v1"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"
)

const grpcPort = 50051

type inventoryService struct {
	inventoryV1.UnimplementedInventoryServiceServer
	mu    sync.RWMutex
	parts map[string]*inventoryV1.Part
}

func (s *inventoryService) GetPart(_ context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	part, ok := s.parts[req.Uuid]
	if !ok {
		return nil, errors.New("part not found")
	}

	response := &inventoryV1.GetPartResponse{
		Info: part,
	}

	return response, nil
}

func (s *inventoryService) ListParts(_ context.Context, req *inventoryV1.ListPartsRequest) (*inventoryV1.ListPartsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	response := &inventoryV1.ListPartsResponse{
		Parts: s.parts,
	}
	return response, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("Ошибка прослушивания: %v\n", err)
		return
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil {
			log.Printf("Ошибка закрытия слушателя: %v\n", cerr)
		}
	}()

	// Создаём GRPc сервер
	s := grpc.NewServer()

	service := &inventoryService{
		parts: getAllMockParts(),
	}

	inventoryV1.RegisterInventoryServiceServer(s, service)

	// Включаем рефлексию для отладки
	reflection.Register(s)

	go func() {
		log.Printf("🚀 gRPC слушаем на порту %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("Ошибка при работе с сервером: %v\n", err)
			return
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("🛑 Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("✅ Server stopped")
}

func getAllMockParts() map[string]*inventoryV1.Part {
	parts := make(map[string]*inventoryV1.Part)

	part := &inventoryV1.Part{
		Uuid:     "a0ad507d-2b70-49e4-9378-3d92ebf9e405",
		Name:     "Двигатель",
		Category: inventoryV1.Category_CATEGORY_ENGINE,
		Manufacturer: &inventoryV1.Manufacturer{
			Name:    "Корпорация двигателей",
			Country: "Россия",
		},
		Tags:          []string{"двигатель", "Россия"},
		Price:         14250000,
		StockQuantity: 7,
	}

	parts[part.Uuid] = part

	return parts
}
