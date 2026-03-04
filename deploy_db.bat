@echo off
REM ============================================
REM vpublish 数据库部署脚本 (Windows)
REM ============================================

setlocal enabledelayedexpansion

REM 配置数据库连接信息
set DB_HOST=localhost
set DB_PORT=3306
set DB_USER=root
set DB_PASS=
set DB_NAME=vpublish

REM 解析命令行参数
:parse_args
if "%~1"=="" goto end_parse
if /i "%~1"=="-h" set DB_HOST=%~2& shift & shift & goto parse_args
if /i "%~1"=="--host" set DB_HOST=%~2& shift & shift & goto parse_args
if /i "%~1"=="-P" set DB_PORT=%~2& shift & shift & goto parse_args
if /i "%~1"=="--port" set DB_PORT=%~2& shift & shift & goto parse_args
if /i "%~1"=="-u" set DB_USER=%~2& shift & shift & goto parse_args
if /i "%~1"=="--user" set DB_USER=%~2& shift & shift & goto parse_args
if /i "%~1"=="-p" set DB_PASS=%~2& shift & shift & goto parse_args
if /i "%~1"=="--password" set DB_PASS=%~2& shift & shift & goto parse_args
echo 未知参数: %~1
goto show_help
:end_parse

echo ============================================
echo   vpublish 数据库部署脚本
echo ============================================
echo.
echo 数据库配置:
echo   主机: %DB_HOST%
echo   端口: %DB_PORT%
echo   用户: %DB_USER%
echo   数据库: %DB_NAME%
echo.

REM 检查 mysql 命令是否存在
where mysql >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 mysql 命令，请确保 MySQL 客户端已安装并添加到 PATH
    pause
    exit /b 1
)

REM 检查 SQL 文件是否存在
if not exist "%~dp0migrations\init.sql" (
    echo [错误] 未找到 SQL 文件: %~dp0migrations\init.sql
    pause
    exit /b 1
)

echo 正在部署数据库...
echo.

REM 执行 SQL 脚本
if "%DB_PASS%"=="" (
    mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER% < "%~dp0migrations\init.sql"
) else (
    mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER% -p%DB_PASS% < "%~dp0migrations\init.sql"
)

if %ERRORLEVEL% equ 0 (
    echo.
    echo ============================================
    echo   数据库部署成功!
    echo ============================================
    echo.
    echo 默认管理员账户:
    echo   用户名: admin
    echo   密码: admin123
    echo.
    echo 测试APP密钥:
    echo   AppKey: test_app_key_12345678
    echo   AppSecret: test_app_secret_abcdefgh
    echo.
) else (
    echo.
    echo [错误] 数据库部署失败，请检查配置和权限
)

pause
exit /b %ERRORLEVEL%

:show_help
echo.
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   -h, --host      数据库主机 (默认: localhost)
echo   -P, --port      数据库端口 (默认: 3306)
echo   -u, --user      数据库用户 (默认: root)
echo   -p, --password  数据库密码 (默认: 空)
echo.
echo 示例:
echo   %~nx0
echo   %~nx0 -h 192.168.1.100 -u root -p mypassword
echo.
pause
exit /b 1