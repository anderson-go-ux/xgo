@echo off
IF not "%1"=="" (
   echo  "%1" > version.txt
)
git add *
git commit -m'auto'
git push
IF not  "%1"=="" (
    git tag %1
    git push --tags
)