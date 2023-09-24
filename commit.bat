@echo off
setlocal enabledelayedexpansion
set "param=%~1"
:trimLeading
if "!param:~0,1!"==" " set "param=!param:~1!" & goto :trimLeading
:trimTrailing
if "!param:~-1!"==" " set "param=!param:~0,-1!" & goto :trimTrailing

IF not !param!=="" (
   echo > version.txt !param!
)
git add *
git commit -m'auto'
git push
IF not !param!=="" (
    git tag %1
    git push --tags
)
endlocal