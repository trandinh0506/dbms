package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (s *Service) ProcessFinalReport(courseID int, isSafe bool) (string, []gin.H, error) {
	const procedureName = "FinalizeAcademicReport"
	p1, p2, err := s.Repo.CallFinalize(courseID, isSafe, procedureName)
	if err != nil {
		return "", nil, err
	}

	if len(p1) != len(p2) {
		return "CRITICAL: Số lượng bản ghi thay đổi trong khi xử lý!", p2, nil
	}

	for i := range p1 {
		if p1[i]["mark"] != p2[i]["mark"] || p1[i]["sid"] != p2[i]["sid"] {
			return "WARNING: Dữ liệu bị thay đổi bởi giảng viên khác! Vui lòng thử lại.", p2, nil
		}
	}

	return "SUCCESS: Báo cáo đã chốt dữ liệu an toàn.", p1, nil
}

func (s *Service) ProcessWeakStudentReport(courseID int, isSafe bool) (string, []gin.H, error) {
	const procedureName = "StatWeakStudentsReport"
	p1, p2, err := s.Repo.CallFinalize(courseID, isSafe, procedureName)
	if err != nil {
		return "", nil, err
	}

	if len(p2) > len(p1) {
		msg := fmt.Sprintf("PHANTOM DETECTED: Ban đầu có %d học viên yếu, nhưng hiện tại đã lên tới %d học viên!", len(p1), len(p2))
		return msg, p2, nil
	}

	if len(p2) < len(p1) {
		msg := fmt.Sprintf("PHANTOM DETECTED: Một số học viên vừa được xóa khỏi danh sách yếu! (%d -> %d)", len(p1), len(p2))
		return msg, p2, nil
	}

	if len(p1) == 0 {
		return "SUCCESS: Không có học viên yếu nào trong lớp.", p1, nil
	}

	return "SUCCESS: Thống kê học viên yếu chính xác.", p1, nil
}
