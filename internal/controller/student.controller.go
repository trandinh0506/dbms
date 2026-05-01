package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetMyGrades(c *gin.Context) {
	sID := c.MustGet("user_id").(int)

	grades, err := ctrl.Svc.Repo.GetStudentGrades(sID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	gpa, _ := ctrl.Svc.Repo.GetGPA(sID)

	c.JSON(200, gin.H{
		"grades": grades,
		"gpa":    gpa,
	})
}

func (ctrl *Controller) Enroll(c *gin.Context) {
	sID := c.MustGet("user_id").(int)

	var req struct {
		CourseID int `json:"course_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Thiếu mã khóa học"})
		return
	}

	err := ctrl.Svc.Repo.EnrollCourse(sID, req.CourseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký môn học thành công!"})
}

func (ctrl *Controller) GetGradesUnsafe(c *gin.Context) {
	sID := c.MustGet("user_id").(int)

	courses, err := ctrl.Svc.Repo.GetGradesUnsafe(c, sID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"grades": courses})
}

func (ctrl *Controller) GetGradesSafe(c *gin.Context) {
	sID := c.MustGet("user_id").(int)

	mark, err := ctrl.Svc.GetStudentGradesSafe(sID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"grades": mark})
}
