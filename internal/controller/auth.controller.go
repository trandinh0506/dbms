package controller

import (
	"database/sql"
	"dbms/internal/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (ctrl *Controller) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var id int
	var role, hashedPassword string

	err := ctrl.Svc.Repo.DB.QueryRow("SELECT id, role, password_hash FROM ACCOUNT WHERE username=@p1",
		req.Username).Scan(&id, &role, &hashedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống"})
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mật khẩu không chính xác"})
		return
	}

	expirationTime := time.Now().Add(240 * time.Hour)
	claims := &models.UserClaims{
		UserID:   id,
		Username: req.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(ctrl.Config.JWTSecret)
	log.Printf("token: ", tokenString)
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"role":  role,
		"user":  req.Username,
	})
}

func (ctrl *Controller) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Fullname string `json:"fullname" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập đầy đủ thông nil"})
		return
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý mật khẩu"})
		return
	}

	var newID int
	err = ctrl.Svc.Repo.DB.QueryRow("EXEC RegisterAccount @p1, @p2, @p3, @p4",
		req.Username, string(hashedBytes), req.Fullname, req.Role).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đăng ký: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Đăng ký tài khoản thành công",
		"id":      newID,
		"user":    req.Username,
		"role":    req.Role,
	})
}
