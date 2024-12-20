package usecases

import (
	"context"
	"learned-api/domain"
	"learned-api/domain/dtos"
	"time"
)

type StudyGroupUsecase struct {
	sgRepository   domain.StudyGroupRepository
	authRepository domain.AuthRepository
}

func NewStudyGroupUsecase(sgRepository domain.StudyGroupRepository, authRepository domain.AuthRepository) *StudyGroupUsecase {
	return &StudyGroupUsecase{
		sgRepository:   sgRepository,
		authRepository: authRepository,
	}
}

func (usecase *StudyGroupUsecase) CreateStudyGroup(c context.Context, creatorID string, studyGroup domain.StudyGroup) domain.CodedError {
	id, err := usecase.sgRepository.ParseID(creatorID)
	if err != nil {
		return err
	}

	newSG := domain.StudyGroup{
		Name:       studyGroup.Name,
		CourseName: studyGroup.CourseName,
		Owner:      id,
		Posts:      []domain.Post{},
	}

	if err := usecase.sgRepository.CreateStudyGroup(c, id, newSG); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) DeleteStudyGroup(c context.Context, studentID string, studyGroupID string) domain.CodedError {
	foundSG, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	if usecase.sgRepository.StringifyID(foundSG.Owner) != studentID {
		return domain.NewError("only the original owner can delete the study group", domain.ERR_FORBIDDEN)
	}

	if err = usecase.sgRepository.DeleteStudyGroup(c, studyGroupID); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) GetPosts(c context.Context, tokenID string, studyGroupID string) ([]domain.GetPostDTO, domain.CodedError) {
	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return []domain.GetPostDTO{}, err
	}

	allowed := false
	for _, sID := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(sID) == tokenID {
			allowed = true
			break
		}
	}

	if !allowed {
		return []domain.GetPostDTO{}, domain.NewError("only students added to the study group can get posts", domain.ERR_FORBIDDEN)
	}

	res := []domain.GetPostDTO{}
	for _, post := range studyGroup.Posts {
		postDto := domain.GetPostDTO{
			Data: post,
		}

		user, err := usecase.authRepository.GetUserByID(c, usecase.sgRepository.StringifyID(post.CreatorID))
		if err != nil {
			postDto.CreatorName = usecase.sgRepository.StringifyID(post.CreatorID)
		} else {
			postDto.CreatorName = user.Name
		}

		res = append(res, postDto)
	}

	return res, nil
}

func (usecase *StudyGroupUsecase) AddPost(c context.Context, creatorID string, studyGroupID string, post domain.Post) domain.CodedError {
	if post.Content == "" {
		return domain.NewError("post content cannot be empty", domain.ERR_BAD_REQUEST)
	}

	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	allowed := false
	for _, studentID := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(studentID) == creatorID {
			allowed = true
			break
		}
	}

	if !allowed {
		return domain.NewError("only students added to the study group can add posts", domain.ERR_FORBIDDEN)
	}

	cID, err := usecase.sgRepository.ParseID(creatorID)
	if err != nil {
		return err
	}

	post.Comments = []domain.Comment{}
	post.CreatedAt = time.Now().Round(0)
	post.CreatorID = cID
	if err = usecase.sgRepository.AddPost(c, studyGroupID, post); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) UpdatePost(c context.Context, creatorID string, studyGroupID string, postID string, post dtos.UpdatePostDTO) domain.CodedError {
	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	allowed := false
	for _, studyGroupID := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(studyGroupID) == creatorID {
			allowed = true
			break
		}
	}

	if !allowed {
		return domain.NewError("only students added to the study group can update posts", domain.ERR_FORBIDDEN)
	}

	if err = usecase.sgRepository.UpdatePost(c, studyGroupID, postID, post); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) RemovePost(c context.Context, creatorID string, studyGroupID string, postID string) domain.CodedError {
	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	allowed := false
	for _, studentID := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(studentID) == creatorID {
			allowed = true
			break
		}
	}

	if !allowed {
		return domain.NewError("only students added to the study group can remove posts", domain.ERR_FORBIDDEN)
	}

	if err = usecase.sgRepository.RemovePost(c, studyGroupID, postID); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) AddComment(c context.Context, creatorID string, studyGroupID string, postID string, comment domain.Comment) domain.CodedError {
	if comment.Content == "" {
		return domain.NewError("comment content cannot be empty", domain.ERR_BAD_REQUEST)
	}

	id, err := usecase.sgRepository.ParseID(creatorID)
	if err != nil {
		return err
	}

	foundUser, err := usecase.authRepository.GetUserByID(c, creatorID)
	if err != nil {
		return err
	}

	comment.CreatedAt = time.Now().Round(0)
	comment.CreatorID = id
	comment.CreatorName = foundUser.Name
	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	allowed := false
	for _, studentID := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(studentID) == creatorID {
			allowed = true
			break
		}
	}

	if !allowed {
		for _, studentID := range studyGroup.Students {
			if usecase.sgRepository.StringifyID(studentID) == creatorID {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		return domain.NewError("only students added to the study group can add comments", domain.ERR_FORBIDDEN)
	}

	if err = usecase.sgRepository.AddComment(c, studyGroupID, postID, comment); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) RemoveComment(c context.Context, userID string, studyGroupID string, postID string, commentID string) domain.CodedError {
	_, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	post, err := usecase.sgRepository.FindPost(c, studyGroupID, postID)
	if err != nil {
		return err
	}

	found := false
	for _, comment := range post.Comments {
		if usecase.sgRepository.StringifyID(comment.ID) == commentID {
			if usecase.sgRepository.StringifyID(comment.CreatorID) != userID {
				return domain.NewError("only the creator of the comment can remove it", domain.ERR_FORBIDDEN)
			}
			found = true
			break
		}
	}

	if !found {
		return domain.NewError("comment not found", domain.ERR_NOT_FOUND)
	}

	if err = usecase.sgRepository.RemoveComment(c, studyGroupID, postID, commentID); err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) AddStudent(c context.Context, tokenID string, studentEmail string, studyGroupID string) domain.CodedError {
	foundUser, err := usecase.authRepository.GetUserByEmail(c, studentEmail)
	if err != nil {
		return err
	}

	if foundUser.Type == domain.RoleTeacher {
		return domain.NewError("can not add teachers as students", domain.ERR_BAD_REQUEST)
	}

	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	allowed := false
	for _, teacher := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(teacher) == tokenID {
			allowed = true
			break
		}
	}

	if !allowed {
		return domain.NewError("only students added to the classroom can add students", domain.ERR_FORBIDDEN)
	}

	targetID := usecase.sgRepository.StringifyID(foundUser.ID)
	found := false
	for _, student := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(student) == targetID {
			found = true
			break
		}
	}

	if found {
		return domain.NewError("student has already been added to the study group", domain.ERR_BAD_REQUEST)
	}

	err = usecase.sgRepository.AddStudent(c, targetID, studyGroupID)
	if err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) RemoveStudent(c context.Context, tokenID string, studyGroupID string, studentID string) domain.CodedError {
	foundUser, err := usecase.authRepository.GetUserByID(c, studentID)
	if err != nil {
		return err
	}

	if foundUser.Type == domain.RoleTeacher {
		return domain.NewError("teachers do not have access to study groups", domain.ERR_FORBIDDEN)
	}

	studyGroup, err := usecase.sgRepository.FindStudyGroup(c, studyGroupID)
	if err != nil {
		return err
	}

	allowed := false
	if usecase.sgRepository.StringifyID(studyGroup.Owner) == tokenID {
		allowed = true
	}

	if usecase.sgRepository.StringifyID(studyGroup.Owner) == studentID {
		return domain.NewError("the owner can not remove themselves from the classroom", domain.ERR_BAD_REQUEST)
	}

	if !allowed {
		return domain.NewError("only the owner of the classroom can remove students", domain.ERR_FORBIDDEN)
	}

	targetID := usecase.sgRepository.StringifyID(foundUser.ID)
	found := false
	for _, student := range studyGroup.Students {
		if usecase.sgRepository.StringifyID(student) == targetID {
			found = true
			break
		}
	}

	if !found {
		return domain.NewError("student is not in the study group", domain.ERR_BAD_REQUEST)
	}

	err = usecase.sgRepository.RemoveStudent(c, targetID, studyGroupID)
	if err != nil {
		return err
	}

	return nil
}

func (usecase *StudyGroupUsecase) GetStudyGroups(c context.Context, tokenID string) ([]domain.StudyGroup, domain.CodedError) {
	foundUser, err := usecase.authRepository.GetUserByID(c, tokenID)
	if err != nil {
		return []domain.StudyGroup{}, err
	}

	studyGroups, err := usecase.sgRepository.GetStudyGroups(c, usecase.sgRepository.StringifyID(foundUser.ID))
	if err != nil {
		return []domain.StudyGroup{}, err
	}

	if len(studyGroups) == 0 {
		return []domain.StudyGroup{}, nil
	}

	return studyGroups, nil
}
