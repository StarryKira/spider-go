package repository

import "gorm.io/gorm"

type NoticeWriter interface {
}

type NoticeReader interface {
}

type NoticeRepository interface {
}

type gormNoticeRepository struct {
	gormDB *gorm.DB
}

func GetAllNotice() {

}
