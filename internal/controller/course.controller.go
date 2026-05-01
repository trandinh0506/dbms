package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetAllCourses(c *gin.Context) {
	courses, err := ctrl.Svc.Repo.GetAllCourses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách khóa học"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(courses),
		"data":  courses,
	})
}
