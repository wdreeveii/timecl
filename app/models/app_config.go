package models

import (
	"fmt"
	"github.com/coopernurse/gorp"
	"strconv"
	"timecl/app/logger"
)

type AppConfig struct {
	ConfigId int
	Key      string
	Val      string
}

func (c AppConfig) String() string {
	return fmt.Sprintf("Config(%s)", c.Key)
}

type EmailSettingsProvider struct {
}

func (m EmailSettingsProvider) ProcessEmailRateLimits(email_settings logger.Email) (int, int, error) {
	var maxmsgs, maxduration int
	if email_settings.MaxMsgs != "" && email_settings.MaxMsgs != "0" {
		if email_settings.MaxDuration != "" && email_settings.MaxDuration != "0" {
			imaxmsgs, err := strconv.ParseInt(email_settings.MaxMsgs, 10, 32)
			if err != nil {
				return 0, 0, err
			}
			maxmsgs = int(imaxmsgs)
			imaxduration, err := strconv.ParseInt(email_settings.MaxDuration, 10, 32)
			if err != nil {
				return 0, 0, err
			}
			maxduration = int(imaxduration)
		}
	}
	return maxmsgs, maxduration, nil
}

func (m EmailSettingsProvider) GetEmail(txn *gorp.Transaction) (logger.Email, error) {
	var email logger.Email
	email_config_data, err := txn.Select(AppConfig{},
		`select * from AppConfig WHERE AppConfig.Key IN ('smtp_address', 
'smtp_port', 'smtp_ssl', 'smtp_username',
'smtp_password', 'smtp_auth_type',
'smtp_max_msgs', 'smtp_max_duration')`)
	if err != nil {
		return logger.Email{}, err
	}

	for _, v := range email_config_data {
		config_row := v.(*AppConfig)
		switch config_row.Key {
		case "smtp_address":
			email.Addr = config_row.Val
		case "smtp_port":
			email.Port = config_row.Val
		case "smtp_ssl":
			email.SSL = config_row.Val
		case "smtp_username":
			email.Username = config_row.Val
		case "smtp_password":
			email.Password = config_row.Val
		case "smtp_auth_type":
			email.AuthType = config_row.Val
		case "smtp_max_msgs":
			email.MaxMsgs = config_row.Val
		case "smtp_max_duration":
			email.MaxDuration = config_row.Val
		}
	}
	return email, nil
}

func SaveEmail(txn *gorp.Transaction, email logger.Email) error {
	var q string = "REPLACE INTO AppConfig (Key, Val) VALUES "
	q += "('smtp_address','" + email.Addr + "'),"
	q += "('smtp_port','" + email.Port + "'),"
	q += "('smtp_ssl','" + email.SSL + "'),"
	q += "('smtp_username', '" + email.Username + "'),"
	q += "('smtp_password', '" + email.Password + "'),"
	q += "('smtp_auth_type', '" + email.AuthType + "'),"
	q += "('smtp_max_msgs', '" + email.MaxMsgs + "'),"
	q += "('smtp_max_duration', '" + email.MaxDuration + "')"
	_, err := txn.Exec(q)
	return err
}
