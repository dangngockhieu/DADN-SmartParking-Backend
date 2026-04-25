package user

import (
	"strings"

	appErrors "backend/internal/common/errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FindWithPagination(page, limit int, search string) (*UserPaginationResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	users, total, err := s.repo.FindWithPagination(page, limit, search)
	if err != nil {
		return nil, appErrors.NewInternal("Lấy danh sách người dùng thất bại")
	}

	result := make([]UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, UserResponse{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Role:      u.Role,
		})
	}

	return &UserPaginationResponse{
		Users: result,
		Total: total,
	}, nil
}

func (s *Service) CreateUserByAdmin(req CreateUserRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	exist, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return appErrors.NewInternal("Kiểm tra email thất bại")
	}
	if exist != nil {
		return appErrors.NewConflict("Email đã được đăng ký!")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternal("Mã hóa mật khẩu thất bại")
	}

	user := &User{
		Email:      req.Email,
		Password:   string(hashed),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Role:       req.Role,
		IsVerified: true,
	}

	if err := s.repo.Create(user); err != nil {
		return appErrors.NewInternal("Tạo người dùng thất bại")
	}

	return nil
}

func (s *Service) ChangePassword(userID uint, req ChangePasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return appErrors.NewBadRequest("Mật khẩu mới và xác nhận mật khẩu không khớp!")
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return appErrors.NewInternal("Lấy thông tin người dùng thất bại")
	}
	if user == nil {
		return appErrors.NewNotFound("Người dùng không tồn tại!")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return appErrors.NewBadRequest("Mật khẩu cũ không đúng!")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternal("Mã hóa mật khẩu thất bại")
	}

	if err := s.repo.UpdatePassword(userID, string(hashed)); err != nil {
		return appErrors.NewInternal("Đổi mật khẩu thất bại")
	}

	return nil
}

func (s *Service) ChangeRole(userID uint, req ChangeRoleRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return appErrors.NewInternal("Lấy thông tin người dùng thất bại")
	}
	if user == nil {
		return appErrors.NewNotFound("Người dùng không tồn tại!")
	}

	if err := s.repo.UpdateRole(userID, req.NewRole); err != nil {
		return appErrors.NewInternal("Đổi vai trò thất bại")
	}

	return nil
}
