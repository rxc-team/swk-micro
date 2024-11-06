@echo off

D:

start cmd /k "title consul &&consul agent -dev"
choice /t 2 /d y /n >nul
start cmd /k "title consul micro web&& micro   web"
start cmd /k "title consul micro api&& micro  --enable_stats api --handler=web"
choice /t 2 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/global
start cmd /k "title global window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/manage
start cmd /k "title manage window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/storage
start cmd /k "title storage window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/database
start cmd /k "title database window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/report
start cmd /k "title report window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/task
start cmd /k "title task window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/import
start cmd /k "title import window&&go run main.go wplugins.go"
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./srv/workflow
start cmd /k "title workflow window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./api/internal
start cmd /k "title internal window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./api/outer
start cmd /k "title outer window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

cd %~dp0
cd..
cd  ./api/system
start cmd /k "title system window&&go run main.go wplugins.go "
choice /t 5 /d y /n >nul

pause