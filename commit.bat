@echo off
git add *
git commit -m'auto'
git push
IF "%1"=="" (
    echo No tag name provided. Exiting...
) ELSE (
    git tag %1
    git push --tags
)
