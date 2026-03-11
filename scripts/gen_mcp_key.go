package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taerc/vpublish/internal/config"
	"github.com/taerc/vpublish/internal/database"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/utils"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	// 自动迁移
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	repo := repository.NewMCPCredentialRepository(db)

	// 生成 AppKey 和 AppSecret
	appKey, err := utils.GenerateAppKey()
	if err != nil {
		log.Fatalf("generate app key: %v", err)
	}

	appSecret, err := utils.GenerateAppSecret()
	if err != nil {
		log.Fatalf("generate app secret: %v", err)
	}

	// 创建凭证
	cred := &model.MCPCredential{
		Name:            "Trae MCP Client",
		AppKey:          appKey,
		AppSecret:       appSecret,
		PermissionLevel: model.PermissionReadWrite,
		Description:     "MCP credential for Trae IDE",
		IsActive:        true,
		CreatedBy:       0,
	}

	if err := repo.Create(context.Background(), cred); err != nil {
		log.Fatalf("create credential: %v", err)
	}

	fmt.Println("==========================================")
	fmt.Println("MCP Credential Created Successfully!")
	fmt.Println("==========================================")
	fmt.Printf("ID:              %d\n", cred.ID)
	fmt.Printf("Name:            %s\n", cred.Name)
	fmt.Printf("Permission:      %s\n", cred.PermissionLevel)
	fmt.Println("------------------------------------------")
	fmt.Printf("MCP_APP_KEY:     %s\n", appKey)
	fmt.Printf("MCP_APP_SECRET:  %s\n", appSecret)
	fmt.Println("==========================================")
	fmt.Println("\nTrae Configuration:")
	fmt.Println(`{
  "mcpServers": [
    {
      "name": "vpublish-mcp",
      "command": ["D:\\wkspace\\git\\vpublish\\vpublish-mcp.exe"],
      "env": {
        "MCP_APP_KEY": "` + appKey + `",
        "MCP_APP_SECRET": "` + appSecret + `"
      }
    }
  ]
}`)
}
