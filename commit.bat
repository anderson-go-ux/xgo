@echo off
IF not "%1"=="" (
    setlocal enabledelayedexpansion
    set "param=%~1"
    :trimLeading
    if "!param:~0,1!"==" " set "param=!param:~1!" & goto :trimLeading
    echo !param! > version.txt
)
git add *
git commit -m'auto'
git push
IF not  "%1"=="" (
    git tag %1
    git push --tags
)