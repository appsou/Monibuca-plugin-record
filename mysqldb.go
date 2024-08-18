package record

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var mysqldb *gorm.DB
var err error

var createDataBaseSql = `CREATE DATABASE IF NOT EXISTS m7srecord;`

var useDataBaseSql = `USE m7srecord;`

var initsql = `CREATE TABLE IF NOT EXISTS eventrecord (
  id int(11) NOT NULL AUTO_INCREMENT,
  streamPath varchar(255) NOT NULL COMMENT '流路径',
  eventId varchar(255) DEFAULT NULL COMMENT '事件编号',
  eventType varchar(255) DEFAULT NULL COMMENT '事件类型',
  eventName varchar(255) DEFAULT NULL COMMENT '事件名称',
  beforeDuration int(255) DEFAULT NULL COMMENT '事件前缓存时长',
  afterDuration int(255) DEFAULT NULL COMMENT '事件后缓存时长',
  recordTime datetime DEFAULT NULL COMMENT '录像时间',
  startTime datetime DEFAULT NULL COMMENT '录像开始时间',
  endTime datetime DEFAULT NULL COMMENT '录像结束时间',
  filepath varchar(255) DEFAULT NULL COMMENT '录像文件路径',
  isDelete varchar(255) DEFAULT '0' COMMENT '是否删除，0表示正常，1表示删除，默认0',
  fileName varchar(255) DEFAULT NULL COMMENT '文件名',
  userId int(11) DEFAULT NULL COMMENT '用户id',
  PRIMARY KEY (id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;`

func initMysqlDB(MysqlDSN string) {
	mysqldb, err = gorm.Open(mysql.Open(MysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	mysqldb.Exec(createDataBaseSql)
	mysqldb.Exec(useDataBaseSql)
	mysqldb.Exec(initsql)
	//mysqldb.AutoMigrate(&EventRecord{})
	mysqldb.AutoMigrate(&Exception{})
}

func paginate[T any](model T, pageNum, pageSize int, filters map[string]interface{}) ([]T, int64, error) {
	var results []T
	var totalCount int64

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 查询总记录数
	countQuery := mysqldb.Model(model)
	for field, value := range filters {
		if valueStr, ok := value.(string); ok && (valueStr != "") {
			//	countQuery = countQuery.Where(field+" LIKE ?", "%"+valueStr+"%")
			//} else {
			if field == "startTime" {
				countQuery = countQuery.Where("recordTime >= ?", valueStr)
			} else if field == "endTime" {
				countQuery = countQuery.Where("recordTime <= ?", valueStr)
			} else {
				countQuery = countQuery.Where(field+" = ?", value)
			}
		}
	}
	result := countQuery.Count(&totalCount)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// 查询当前页的数据
	query := mysqldb.Model(model).Limit(pageSize).Offset(offset)
	for field, value := range filters {
		if valueStr, ok := value.(string); ok && (valueStr != "") {
			//	query = query.Where(field+" LIKE ?", "%"+valueStr+"%")
			//} else {
			if field == "startTime" {
				query = query.Where("recordTime >= ?", valueStr)
			} else if field == "endTime" {
				query = query.Where("recordTime <= ?", valueStr)
			} else {
				query = query.Where(field+" = ?", value)
			}
		}
	}

	result = query.Find(&results)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return results, totalCount, nil
}
