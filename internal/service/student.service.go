package service

import (
	"context"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetStudentGradesSafe(studentID int) ([]gin.H, error) {
	data, err := s.Repo.GetGradesSafe(context.Background(), studentID)
	if err != nil {
		return nil, err
	}

	var result []gin.H
	for _, item := range data {
		result = append(result, gin.H{
			"course_name": item.Name,
			"mark":        item.Mark,
		})
	}
	return result, nil
}
