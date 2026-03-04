package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/taerc/vpublish/internal/config"
	"github.com/taerc/vpublish/internal/database"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/pinyin"
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

	// 初始化 Repository
	categoryRepo := repository.NewCategoryRepository(db)
	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// 创建 MCP 服务器
	s := server.NewMCPServer(
		"vpublish-mcp",
		"1.0.0",
	)

	// 注册工具
	registerTools(s, categoryRepo, packageRepo, versionRepo, statsRepo)

	// 启动服务
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// getArgs 从请求中获取参数 map
func getArgs(request mcp.CallToolRequest) map[string]interface{} {
	if request.Params.Arguments == nil {
		return make(map[string]interface{})
	}
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}
	return make(map[string]interface{})
}

func registerTools(
	s *server.MCPServer,
	categoryRepo *repository.CategoryRepository,
	packageRepo *repository.PackageRepository,
	versionRepo *repository.VersionRepository,
	statsRepo *repository.StatsRepository,
) {
	// 1. 列出所有类别
	s.AddTool(mcp.Tool{
		Name:        "list_categories",
		Description: "获取所有软件类别列表",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		categories, err := categoryRepo.ListActive(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(categories, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	// 2. 列出软件包
	s.AddTool(mcp.Tool{
		Name:        "list_packages",
		Description: "获取软件包列表",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(request)
		categoryID := uint(0)
		if val, ok := args["category_id"].(float64); ok {
			categoryID = uint(val)
		}

		packages, _, err := packageRepo.List(ctx, categoryID, 1, 100)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(packages, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	// 3. 列出软件包版本
	s.AddTool(mcp.Tool{
		Name:        "list_versions",
		Description: "获取软件包的所有版本",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(request)
		packageID, ok := args["package_id"].(float64)
		if !ok {
			return mcp.NewToolResultError("package_id is required"), nil
		}

		versions, _, err := versionRepo.ListByPackage(ctx, uint(packageID), 1, 100)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(versions, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	// 4. 获取最新版本
	s.AddTool(mcp.Tool{
		Name:        "get_latest_version",
		Description: "获取指定类别的最新版本",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(request)
		categoryCode, ok := args["category_code"].(string)
		if !ok {
			return mcp.NewToolResultError("category_code is required"), nil
		}

		version, err := versionRepo.GetLatestByCategoryCode(ctx, categoryCode)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]interface{}{
			"id":            version.ID,
			"version":       version.Version,
			"version_code":  version.VersionCode,
			"file_name":     version.FileName,
			"file_size":     version.FileSize,
			"file_hash":     version.FileHash,
			"changelog":     version.Changelog,
			"force_upgrade": version.ForceUpgrade,
			"is_stable":     version.IsStable,
			"published_at":  version.PublishedAt,
			"package":       version.Package,
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	// 5. 创建类别
	s.AddTool(mcp.Tool{
		Name:        "create_category",
		Description: "创建新的软件类别",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(request)
		name, ok := args["name"].(string)
		if !ok {
			return mcp.NewToolResultError("name is required"), nil
		}

		description, _ := args["description"].(string)

		// 生成拼音代码
		code := pinyin.GenerateCode(name)

		category := &model.Category{
			Name:        name,
			Code:        code,
			Description: description,
			IsActive:    true,
		}

		if err := categoryRepo.Create(ctx, category); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(category, "", "  ")
		return mcp.NewToolResultText("Created category:\n" + string(data)), nil
	})

	// 6. 获取下载统计
	s.AddTool(mcp.Tool{
		Name:        "get_download_stats",
		Description: "获取下载统计数据",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(request)
		statsType, _ := args["type"].(string)
		categoryID := uint(0)
		if val, ok := args["category_id"].(float64); ok {
			categoryID = uint(val)
		}

		var result interface{}
		var err error

		switch statsType {
		case "daily":
			dateStr, _ := args["date"].(string)
			date := parseDate(dateStr)
			var count int64
			count, err = statsRepo.GetDailyStats(ctx, categoryID, date)
			result = map[string]interface{}{
				"type":  "daily",
				"date":  date.Format("2006-01-02"),
				"count": count,
			}
		case "monthly":
			year := int(args["year"].(float64))
			month := int(args["month"].(float64))
			var count int64
			count, err = statsRepo.GetMonthlyStats(ctx, categoryID, year, month)
			result = map[string]interface{}{
				"type":  "monthly",
				"year":  year,
				"month": month,
				"count": count,
			}
		case "yearly":
			year := int(args["year"].(float64))
			var count int64
			count, err = statsRepo.GetYearlyStats(ctx, categoryID, year)
			result = map[string]interface{}{
				"type":  "yearly",
				"year":  year,
				"count": count,
			}
		default:
			return mcp.NewToolResultError("invalid type, must be daily/monthly/yearly"), nil
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	// 7. 删除版本
	s.AddTool(mcp.Tool{
		Name:        "delete_version",
		Description: "删除指定版本",
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(request)
		versionID, ok := args["version_id"].(float64)
		if !ok {
			return mcp.NewToolResultError("version_id is required"), nil
		}

		if err := versionRepo.Delete(ctx, uint(versionID)); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Version %d deleted successfully", int(versionID))), nil
	})
}

// parseDate 解析日期
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}
	date, _ := time.Parse("2006-01-02", dateStr)
	return date
}
