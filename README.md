#### CREATE BY : FLORIAN BATTESTI
#### FOR : Create "low cost dropbox like" with AWS S3 but with encrypt file.
#### START WHEN : 12-2022

## PREREQUISITE

#### Golang install > 1.9 
#### AWS CLI and profile already set up

## How to set up Go 
```
go mod init main
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
go get github.com/aws/aws-sdk-go-v2/service/s3

```
## Set your key and store it ( can be 8 / 16 / 32 charac )
```
config/key.txt
```

## Don't loose your key ! Keep it safe 


## What is done by the script :
#### syncLocalToS3.go : sync local folder and target S3 bucket.

## Command for start script
```
go run syncLocalToS3.go
```




# source doc :
https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html
https://medium.com/@mertkimyonsen/encrypt-a-file-using-go-f1fe3bc7c635
