package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
)

func main() {
	// 解析命令行参数
	ak := flag.String("ak", "GJXLEJRZ7MNXNVWDVLW1", "Access Key")
	sk := flag.String("sk", "g31jsjzedeT6bh9H6zztyCOrdaimzQTRPVgF18De", "Secret Key")
	recordName := flag.String("rm", "w", "Record Name")
	domain := flag.String("dm", "zjlchb.com", "Domain Name")
	url := flag.String("ur", "https://ww.zjlchb.com:34567", "URL for the record")
	flag.Parse()

	*domain = *domain + "."
	*recordName = *recordName + "." + *domain
	*url = "301 " + *url

	// 检查参数是否完整
	if *ak == "" || *sk == "" || *recordName == "" || *domain == "" || *url == "" {
		fmt.Println("请提供所有必要的参数")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 创建认证信息
	auth := basic.NewCredentialsBuilder().
		WithAk(*ak).
		WithSk(*sk).
		Build()

	// 创建DNS客户端
	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-east-3")).
			WithCredential(auth).
			Build())

	// 查询域名的Zone ID
	zoneID, err := getZoneID(client, *domain)
	if err != nil {
		fmt.Println("获取Zone ID失败:", err)
		os.Exit(1)
	}

	// 查询记录名是否存在
	recordID, err := getRecordID(client, zoneID, *recordName)
	if err != nil {
		fmt.Println("查询记录失败:", err)
		os.Exit(1)
	}

	if recordID != "" {
		// 如果记录存在，更新记录
		err = updateRecord(client, zoneID, recordID, *url)
		if err != nil {
			fmt.Println("更新记录失败:", err)
			os.Exit(1)
		}
		fmt.Println("记录更新成功")
	} else {
		// 如果记录不存在，创建记录
		err = createRecord(client, zoneID, *recordName, *url)
		if err != nil {
			fmt.Println("创建记录失败:", err)
			os.Exit(1)
		}
		fmt.Println("记录创建成功")
	}
}

// 在ptr指向的字符串前面追加prefix
func prependString(ptr *string, prefix string) {
	result := prefix + *ptr
	*ptr = result
}

// 在ptr指向的字符串后面追加suffix
func appendString(ptr *string, suffix string) {
	result := *ptr + suffix
	*ptr = result
}

// 获取域名的Zone ID
func getZoneID(client *dns.DnsClient, domain string) (string, error) {
	request := &model.ListPublicZonesRequest{}
	response, err := client.ListPublicZones(request)
	if err != nil {
		return "", err
	}

	// 将响应转换为JSON字符串
	jsonData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("转换响应为JSON失败: %v", err)
	}

	// 解析JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}

	// 提取zones字段
	zones, ok := result["zones"].([]interface{})
	if !ok {
		return "", fmt.Errorf("未找到zones字段")
	}

	// 遍历zones，查找匹配的域名
	for _, zone := range zones {
		zoneMap, ok := zone.(map[string]interface{})
		if !ok {
			continue
		}
		if zoneMap["name"] == domain {
			zoneID, ok := zoneMap["id"].(string)
			if !ok {
				return "", fmt.Errorf("zone ID类型错误")
			}
			return zoneID, nil
		}
	}

	return "", fmt.Errorf("未找到域名 %s 的Zone ID", domain)
}

// 获取记录名的Record ID
func getRecordID(client *dns.DnsClient, zoneID, recordName string) (string, error) {
	request := &model.ListRecordSetsByZoneRequest{
		ZoneId: zoneID,
	}
	response, err := client.ListRecordSetsByZone(request)
	if err != nil {
		return "", err
	}

	// 将响应转换为JSON字符串
	jsonData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("转换响应为JSON失败: %v", err)
	}

	// 解析JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}

	// 提取recordsets字段
	recordsets, ok := result["recordsets"].([]interface{})
	if !ok {
		return "", fmt.Errorf("未找到recordsets字段")
	}

	// 遍历recordsets，查找匹配的记录名
	for _, record := range recordsets {
		recordMap, ok := record.(map[string]interface{})
		if !ok {
			continue
		}
		if recordMap["name"] == recordName {
			recordID, ok := recordMap["id"].(string)
			if !ok {
				return "", fmt.Errorf("record ID类型错误")
			}
			return recordID, nil
		}
	}

	return "", nil
}

// 更新记录
func updateRecord(client *dns.DnsClient, zoneID, recordID, url string) error {
	request := &model.UpdateRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: recordID,
		Body: &model.UpdateRecordSetReq{
			Records: &[]string{url},
		},
	}
	_, err := client.UpdateRecordSet(request)
	return err
}

// 创建记录
func createRecord(client *dns.DnsClient, zoneID, recordName, url string) error {
	request := &model.CreateRecordSetRequest{
		ZoneId: zoneID,
		Body: &model.CreateRecordSetRequestBody{
			Name:    recordName,
			Type:    "REDIRECT_URL",
			Records: []string{url},
		},
	}
	_, err := client.CreateRecordSet(request)
	return err
}
