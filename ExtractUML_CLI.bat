@echo off
setlocal enabledelayedexpansion
set EXE=ExtractorUML.exe

if not exist %EXE% (
    echo Error: %EXE% not found! Please build it first.
    pause
    exit /b 1
)

if "%~1"=="" (
    echo ==================================================
    echo       Interactive UML Detail Extractor
    echo ==================================================
    
    :get_file
    echo.
    set /p "UML_FILE=Nhap ten file .drawio can trich xuat: "
    if not exist "!UML_FILE!" (
        echo [ERROR] Khong tim thay file: !UML_FILE!
        goto get_file
    )

    echo.
    echo Dang trich xuat du lieu tu: !UML_FILE!
    echo --------------------------------------------------
    %EXE% "!UML_FILE!"
    
    echo.
    echo --------------------------------------------------
    pause
    goto get_file
) else (
    %EXE% %*
)
endlocal
