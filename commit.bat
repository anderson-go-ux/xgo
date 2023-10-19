@echo off

set "filePath=version.txt"
set ver=""
for /f "tokens=*" %%a in (%filePath%) do (
    set ver=%%a
)

set ver1=%ver:~0,5%
set ver2=%ver:~5,5%

set /a ver2ex=%ver2% + 1

set ver=%ver1%%ver2ex%


git add *
git commit -m'auto'
git push
git tag %ver%
git push --tags
echo  %ver% > version.txt
