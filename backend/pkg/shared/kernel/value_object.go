package kernel

import (
	"time"
)

// ValueObject 值对象接口
type ValueObject interface {
	Equals(other ValueObject) bool
}

// Email 邮箱值对象
type Email struct {
	Value string `json:"value"`
}

// NewEmail 创建邮箱值对象
func NewEmail(value string) (*Email, error) {
	// 这里可以添加邮箱验证逻辑
	return &Email{Value: value}, nil
}

// Equals 比较邮箱是否相等
func (e *Email) Equals(other ValueObject) bool {
	if otherEmail, ok := other.(*Email); ok {
		return e.Value == otherEmail.Value
	}
	return false
}

// Phone 电话号码值对象
type Phone struct {
	Value string `json:"value"`
}

// NewPhone 创建电话号码值对象
func NewPhone(value string) (*Phone, error) {
	// 这里可以添加电话号码验证逻辑
	return &Phone{Value: value}, nil
}

// Equals 比较电话号码是否相等
func (p *Phone) Equals(other ValueObject) bool {
	if otherPhone, ok := other.(*Phone); ok {
		return p.Value == otherPhone.Value
	}
	return false
}

// Money 金额值对象
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// NewMoney 创建金额值对象
func NewMoney(amount float64, currency string) *Money {
	return &Money{
		Amount:   amount,
		Currency: currency,
	}
}

// Equals 比较金额是否相等
func (m *Money) Equals(other ValueObject) bool {
	if otherMoney, ok := other.(*Money); ok {
		return m.Amount == otherMoney.Amount && m.Currency == otherMoney.Currency
	}
	return false
}

// DateRange 日期范围值对象
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// NewDateRange 创建日期范围值对象
func NewDateRange(startDate, endDate time.Time) (*DateRange, error) {
	if startDate.After(endDate) {
		return nil, ErrInvalidDateRange
	}
	return &DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// Equals 比较日期范围是否相等
func (d *DateRange) Equals(other ValueObject) bool {
	if otherRange, ok := other.(*DateRange); ok {
		return d.StartDate.Equal(otherRange.StartDate) && d.EndDate.Equal(otherRange.EndDate)
	}
	return false
}

// Contains 检查日期是否在范围内
func (d *DateRange) Contains(date time.Time) bool {
	return !date.Before(d.StartDate) && !date.After(d.EndDate)
}

// Duration 获取持续时间
func (d *DateRange) Duration() time.Duration {
	return d.EndDate.Sub(d.StartDate)
}
