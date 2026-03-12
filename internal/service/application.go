package service

import (
    "context"
    "time"

    "github.com/rs/xid"

    "wchat/internal/model"
    "wchat/internal/repository"
    "wchat/pkg/errcode"
)

type ApplicationInfo struct {
    Uuid        string
    UserId      string
    Nickname    string // 动态获取
    Avatar      string // 动态获取
    ContactId   string
    ContactType int8
    Status      int8
    Message     string
    LastApplyAt string
}

type ApplicationService struct {
    applyRepo   *repository.ContactApplyRepo
    userRepo    *repository.UserRepo
    contactRepo *repository.ContactRepo
    // groupRepo *repository.GroupRepo // 预留给处理群聊申请
}

func NewApplicationService(
    applyRepo *repository.ContactApplyRepo,
    userRepo *repository.UserRepo,
    contactRepo *repository.ContactRepo,
) *ApplicationService {
    return &ApplicationService{
        applyRepo:   applyRepo,
        userRepo:    userRepo,
        contactRepo: contactRepo,
    }
}

func (s *ApplicationService) GetApplicationList(ctx context.Context, userID string) ([]ApplicationInfo, error) {
    applies, err := s.applyRepo.FindReceivedApplies(ctx, userID)
    if err != nil {
        return nil, err
    }

    if len(applies) == 0 {
        return []ApplicationInfo{}, nil
    }

    var applicantIDs []string
    for _, app := range applies {
        applicantIDs = append(applicantIDs, app.UserId)
    }

    users, err := s.userRepo.FindByUUIDs(ctx, applicantIDs)
    if err != nil {
        return nil, err
    }

    userMap := make(map[string]model.User)
    for _, u := range users {
        userMap[u.Uuid] = u
    }

    infos := make([]ApplicationInfo, 0, len(applies))
    for _, app := range applies {
        applicant := userMap[app.UserId]
        infos = append(
            infos, ApplicationInfo{
                Uuid:        app.Uuid,
                UserId:      app.UserId,
                Nickname:    applicant.Nickname,
                Avatar:      applicant.Avatar,
                ContactId:   app.ContactId,
                ContactType: app.ContactType,
                Status:      app.Status,
                Message:     app.Message,
                LastApplyAt: app.LastApplyAt.Format(time.RFC3339),
            },
        )
    }

    return infos, nil
}

func (s *ApplicationService) SubmitApplication(
    ctx context.Context, applicantID, contactID string, contactType int8, message string,
) error {
    if applicantID == contactID {
        return errcode.New(errcode.CannotAddSelf) // 不能添加自己
    }

    if contactType == 0 {
        // 检查是否已经是好友
        isFriend, err := s.contactRepo.CheckContactExists(ctx, applicantID, contactID)
        if err != nil {
            return err
        }
        if isFriend {
            return errcode.New(errcode.AlreadyFriends)
        }

        // TODO: 检查对方是否把你拉黑了 (TargetBlocked)
    }

    // 查找是否已经有过申请记录
    apply, err := s.applyRepo.GetApplyByUserAndTarget(ctx, applicantID, contactID, contactType)
    if err != nil {
        return err
    }

    if apply != nil {
        // 如果记录存在且已经是待处理状态，避免重复提交
        if apply.Status == 0 {
            return errcode.NewWithMsg(errcode.ParamError, "您的申请正在处理中，请耐心等待")
        }
        // 如果被拒绝过，允许再次发起，更新状态和时间
        apply.Status = 0
        apply.Message = message
        apply.LastApplyAt = time.Now()
        return s.applyRepo.SaveApply(ctx, apply)
    }

    // 创建全新申请
    newApply := &model.ContactApply{
        Uuid:        xid.New().String(),
        UserId:      applicantID,
        ContactId:   contactID,
        ContactType: contactType,
        Status:      0,
        Message:     message,
        LastApplyAt: time.Now(),
    }
    return s.applyRepo.SaveApply(ctx, newApply)
}

// HandleApplication 处理申请 (同意/拒绝/拉黑)
func (s *ApplicationService) HandleApplication(ctx context.Context, operatorID, applyUUID string, status int8) error {
    apply, err := s.applyRepo.FindApplyByUUID(ctx, applyUUID)
    if err != nil {
        return errcode.New(errcode.ApplyNotFound)
    }

    // 本人操作
    if apply.ContactType == 0 && apply.ContactId != operatorID {
        return errcode.New(errcode.Unauthorized)
    }

    // 幂等校验
    if apply.Status != 0 {
        return errcode.New(errcode.ApplyAlreadyHandled)
    }

    // 单聊处理逻辑
    if apply.ContactType == 0 {
        // 确认两人还没成为好友
        isFriend, err := s.contactRepo.CheckContactExists(ctx, apply.UserId, apply.ContactId)
        if err != nil {
            return err
        }
        if isFriend && status == 1 {
            _ = s.applyRepo.HandleUserApplyTx(ctx, applyUUID, apply.UserId, apply.ContactId, 1)
            return errcode.New(errcode.AlreadyFriends)
        }

        // 调用事务：改状态 + 添好友
        return s.applyRepo.HandleUserApplyTx(ctx, applyUUID, apply.UserId, apply.ContactId, status)
    }

    // TODO: 如果 apply.ContactType == 1 (群聊申请)，这里应调用 GroupRepo 的事务
    return nil
}
