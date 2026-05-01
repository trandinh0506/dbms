package service

import (
	"dbms/internal/repository"

	"github.com/gin-gonic/gin"
)

type Service struct {
	Repo *repository.Repository
	Ssm  *SessionManager
}

func (s *Service) GetCourseStudents(courseID int) ([]gin.H, error) {
	return s.Repo.GetStudentsByCourse(courseID)
}
