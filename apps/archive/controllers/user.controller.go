package controllers

import (
	"github.com/Arxtect/Einstein/apps/archive/models"
	"github.com/Arxtect/Einstein/apps/archive/service/ws"
	"github.com/Arxtect/Einstein/common/constants"
	"github.com/Arxtect/Einstein/common/logger"
	"github.com/Arxtect/Einstein/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type UserController struct {
	DB *gorm.DB
	cs *CreditSystem
}

func NewUserController(DB *gorm.DB, creditSystem *CreditSystem) UserController {
	return UserController{DB, creditSystem}
}

func (uc *UserController) GetMe(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	userResponse := &models.UserResponse{
		ID:        currentUser.ID,
		Name:      currentUser.Name,
		Email:     currentUser.Email,
		Photo:     currentUser.Photo,
		Role:      currentUser.Role,
		Verified:  currentUser.Verified,
		Balance:   currentUser.Balance,
		Provider:  currentUser.Provider,
		CreatedAt: currentUser.CreatedAt,
		UpdatedAt: currentUser.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": userResponse}})
}

func (uc *UserController) AdminUpdateBalance(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "Only Admin can update credits"})
		return
	}
	var payload *models.UpdateBalanceInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	newBalance, err := uc.cs.UpdateBalanceByUserEmail(payload.Email, payload.Amount)
	if err != nil {
		logger.Warning(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Failed to update credits"})
		return
	}
	now := time.Now()
	bill := models.BillingHistory{
		Operator:          currentUser.Email,
		AccountEmail:      payload.Email,
		Amount:            payload.Amount,
		TransactionType:   constants.TransactionTypeAdmin,
		TransactionDetail: "Admin updates balance",
		TransactionTime:   now,
	}
	err = uc.DB.Create(&bill).Error
	if err != nil {
		logger.Warning("Failed to create billing history", err)
	}
	logger.Info("Create billing history", bill)

	var user models.User
	err = uc.DB.First(&user, "email = ?", payload.Email).Error
	if err != nil {
		logger.Warning(err.Error())
	}
	var firstName = user.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	emailData := utils.EmailData{
		FirstName: firstName,
		Subject:   "Pointer.ai 充值提醒",
		Amount:    payload.Amount,
		Balance:   newBalance,
	}
	go func() {
		utils.SendEmail(&user, &emailData, "chargeSucceed.html")

	}()
	logger.Info("Charge balance of user", payload.Email, "for amount", payload.Amount, "newBalance: ", newBalance)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"updated_email": payload.Email, "new_balance": newBalance}})
}

// WsEditingRoom 用户socket用来控制接收和写入,达到协同编辑
func (uc *UserController) WsEditingRoom(c *gin.Context) {
	ws.HandlerWs(c)

}

// CreateRoom 创建房间
func (uc *UserController) CreateRoom(c *gin.Context) {
	fileId := c.Param("fileId")
	room, Invitation, err := ws.HandleCreateRoom(fileId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": room, "invitation": Invitation})
}

// WsSubscribeEdit 用户socket用来控制接收和写入,达到协同编辑
//func (uc *UserController) WsSubscribeEdit(c *gin.Context) {
//	currentUser := c.MustGet("currentUser").(models.User)
//
//	var req *models.SubscribeRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
//		return
//	}
//	if err := ws.Subscriber.Subscribe(req.RoomId, subscription{Connection: c.Writer, UserInfo: &currentUser}); err != nil {
//		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
//}
