package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// CreditInfoClient 企业信用信息查询客户端
type CreditInfoClient struct {
	BaseURL    string
	Username   string
	Password   string
	AESKey     string
	RSAPubKey  string
	RSAPriKey  string
	HTTPClient *http.Client
}

// NewCreditInfoClient 创建企业信用信息查询客户端
func NewCreditInfoClient(baseURL, username, password, aesKey, rsaPubKey, rsaPriKey string) *CreditInfoClient {
	return &CreditInfoClient{
		BaseURL:    baseURL,
		Username:   username,
		Password:   password,
		AESKey:     aesKey,
		RSAPubKey:  rsaPubKey,
		RSAPriKey:  rsaPriKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreditInfoRequest 信用信息查询请求
type CreditInfoRequest struct {
	ProductCode   string               `json:"productCode"`
	QueryReasonID int                  `json:"queryReasonId"`
	Conditions    CreditInfoConditions `json:"conditions"`
}

// CreditInfoConditions 查询条件
type CreditInfoConditions struct {
	CodeType       int    `json:"codeType"`
	EnterpriseCode string `json:"enterpriseCode"`
	EnterpriseName string `json:"enterpriseName"`
}

// CreditInfoResponse 信用信息查询响应
type CreditInfoResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    CreditInfoData `json:"data"`
	Error   string         `json:"error,omitempty"`
}

// CreditInfoData 信用信息数据
type CreditInfoData struct {
	SecretKey string `json:"secretKey"`
	Content   string `json:"content"`
}

// CreditInfo 企业信用信息
type CreditInfo struct {
	CompanyName       string    `json:"company_name"`
	CompanyCode       string    `json:"company_code"`
	CreditLevel       string    `json:"credit_level"`
	RiskLevel         string    `json:"risk_level"`
	ComplianceStatus  string    `json:"compliance_status"`
	BusinessStatus    string    `json:"business_status"`
	LegalPerson       string    `json:"legal_person"`
	RegisteredCapital string    `json:"registered_capital"`
	FoundedDate       string    `json:"founded_date"`
	Industry          string    `json:"industry"`
	Address           string    `json:"address"`
	LastUpdated       time.Time `json:"last_updated"`
	// 其他信用信息字段
	RiskFactors     []string `json:"risk_factors,omitempty"`
	CreditScore     int      `json:"credit_score,omitempty"`
	ComplianceItems []string `json:"compliance_items,omitempty"`
}

// GetCompanyCreditInfo 获取企业信用信息
func (c *CreditInfoClient) GetCompanyCreditInfo(companyName, companyCode string) (*CreditInfo, error) {
	// 1. 构建查询参数
	request := CreditInfoRequest{
		ProductCode:   "CR1001",
		QueryReasonID: 1,
		Conditions: CreditInfoConditions{
			CodeType:       1,
			EnterpriseCode: companyCode,
			EnterpriseName: companyName,
		},
	}

	// 2. 序列化请求参数
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求参数失败: %v", err)
	}

	// 3. AES加密请求参数
	encryptedContent, err := c.encryptAES(string(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("AES加密失败: %v", err)
	}

	// 4. RSA加密AES密钥
	encryptedKey, err := c.encryptRSA(c.AESKey)
	if err != nil {
		return nil, fmt.Errorf("RSA加密失败: %v", err)
	}

	// 5. 构建HTTP请求
	formData := map[string]string{
		"secretKey": encryptedKey,
		"content":   encryptedContent,
		"version":   "1.0",
	}

	// 6. 发送HTTP请求
	response, err := c.sendRequest(formData)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 7. 解密响应数据
	creditInfo, err := c.decryptResponse(response)
	if err != nil {
		return nil, fmt.Errorf("解密响应失败: %v", err)
	}

	return creditInfo, nil
}

// encryptAES AES加密
func (c *CreditInfoClient) encryptAES(plaintext string) (string, error) {
	// 创建AES cipher
	block, err := aes.NewCipher([]byte(c.AESKey))
	if err != nil {
		return "", err
	}

	// 填充数据到块大小的倍数
	paddedText := c.pkcs7Padding([]byte(plaintext), aes.BlockSize)

	// 创建ECB模式的加密器
	mode := cipher.NewCBCEncrypter(block, make([]byte, aes.BlockSize))

	// 加密数据
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptAES AES解密
func (c *CreditInfoClient) decryptAES(ciphertext string) (string, error) {
	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 创建AES cipher
	block, err := aes.NewCipher([]byte(c.AESKey))
	if err != nil {
		return "", err
	}

	// 创建ECB模式的解密器
	mode := cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize))

	// 解密数据
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)

	// 去除填充
	unpaddedText := c.pkcs7UnPadding(plaintext)
	return string(unpaddedText), nil
}

// encryptRSA RSA加密
func (c *CreditInfoClient) encryptRSA(plaintext string) (string, error) {
	// 解析RSA公钥
	pubKey, err := c.parseRSAPublicKey(c.RSAPubKey)
	if err != nil {
		return "", err
	}

	// 加密数据
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, []byte(plaintext), nil)
	if err != nil {
		return "", err
	}

	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptRSA RSA解密
func (c *CreditInfoClient) decryptRSA(ciphertext string) (string, error) {
	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 解析RSA私钥
	priKey, err := c.parseRSAPrivateKey(c.RSAPriKey)
	if err != nil {
		return "", err
	}

	// 解密数据
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priKey, data, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// parseRSAPublicKey 解析RSA公钥
func (c *CreditInfoClient) parseRSAPublicKey(pubKeyStr string) (*rsa.PublicKey, error) {
	// 解码Base64
	keyBytes, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, err
	}

	// 解析PEM格式
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		// 如果PEM解析失败，尝试直接使用Base64解码
		decodedKey, err := base64.StdEncoding.DecodeString(c.RSAPubKey)
		if err != nil {
			return nil, fmt.Errorf("无法解析RSA公钥: %v", err)
		}
		block = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: decodedKey,
		}
	}

	// 解析公钥
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA公钥")
	}

	return rsaPubKey, nil
}

// parseRSAPrivateKey 解析RSA私钥
func (c *CreditInfoClient) parseRSAPrivateKey(priKeyStr string) (*rsa.PrivateKey, error) {
	// 解码Base64
	keyBytes, err := base64.StdEncoding.DecodeString(priKeyStr)
	if err != nil {
		return nil, err
	}

	// 解析PEM格式
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("无法解析PEM格式的私钥")
	}

	// 解析私钥
	priKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPriKey, ok := priKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA私钥")
	}

	return rsaPriKey, nil
}

// sendRequest 发送HTTP请求
func (c *CreditInfoClient) sendRequest(formData map[string]string) (*CreditInfoResponse, error) {
	// 构建表单数据
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range formData {
		writer.WriteField(key, value)
	}
	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.BaseURL, &buf)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", c.getBasicAuth())

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var response CreditInfoResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// decryptResponse 解密响应数据
func (c *CreditInfoClient) decryptResponse(response *CreditInfoResponse) (*CreditInfo, error) {
	// 1. RSA解密响应密钥
	decryptedKey, err := c.decryptRSA(response.Data.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("解密响应密钥失败: %v", err)
	}

	// 2. AES解密响应内容
	decryptedContent, err := c.decryptAESWithKey(response.Data.Content, decryptedKey)
	if err != nil {
		return nil, fmt.Errorf("解密响应内容失败: %v", err)
	}

	// 3. 解析信用信息
	var creditInfo CreditInfo
	if err := json.Unmarshal([]byte(decryptedContent), &creditInfo); err != nil {
		return nil, fmt.Errorf("解析信用信息失败: %v", err)
	}

	creditInfo.LastUpdated = time.Now()
	return &creditInfo, nil
}

// decryptAESWithKey 使用指定密钥进行AES解密
func (c *CreditInfoClient) decryptAESWithKey(ciphertext, key string) (string, error) {
	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 创建AES cipher
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// 创建ECB模式的解密器
	mode := cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize))

	// 解密数据
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)

	// 去除填充
	unpaddedText := c.pkcs7UnPadding(plaintext)
	return string(unpaddedText), nil
}

// getBasicAuth 获取Basic认证头
func (c *CreditInfoClient) getBasicAuth() string {
	auth := c.Username + ":" + c.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// pkcs7Padding PKCS7填充
func (c *CreditInfoClient) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7UnPadding PKCS7去填充
func (c *CreditInfoClient) pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
