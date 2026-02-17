package routes

import (
	"refina-transaction/config/miniofs"
	"refina-transaction/interface/grpc/client"
	"refina-transaction/interface/http/handler"
	"refina-transaction/interface/http/middleware"
	"refina-transaction/internal/repository"
	"refina-transaction/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TransactionRoutes(version *gin.Engine, db *gorm.DB, minio *miniofs.MinIOManager) {
	txManager := repository.NewTxManager(db)
	transactionRepo := repository.NewTransactionRepository(db)
	walletRepo := client.NewWalletClient(client.GetManager().GetWalletClient())
	categoryRepo := repository.NewCategoryRepository(db)
	attachmentRepo := repository.NewAttachmentsRepository(db)
	outboxRepository := repository.NewOutboxRepository(db)

	Transaction_serv := service.NewTransactionService(txManager, transactionRepo, walletRepo, categoryRepo, attachmentRepo, outboxRepository, minio)
	Transaction_handler := handler.NewTransactionHandler(Transaction_serv)

	transaction := version.Group("/transactions")
	transaction.Use(middleware.AuthMiddleware())

	transaction.GET("", Transaction_handler.GetAllTransactions)
	transaction.GET(":id", Transaction_handler.GetTransactionByID)
	transaction.GET("user", Transaction_handler.GetTransactionsByUserID)
	transaction.POST(":type", Transaction_handler.CreateTransaction)
	transaction.POST("attachment/:id", Transaction_handler.UploadAttachment)
	transaction.PUT(":id", Transaction_handler.UpdateTransaction)
	transaction.DELETE(":id", Transaction_handler.DeleteTransaction)
}
