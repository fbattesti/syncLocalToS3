package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Custom error management because i don't like the Go best practice
func checkError(err error, message string) int {
	if err != nil {
		log.Fatalln("ERROR IN SECTION : ", message, "\nERROR = ", err)
		return 1
	}
	return 0
}

func listFileInBucketS3(bucketName string, awsProfile string) []string {

	// slice-array where i will store the result
	sliceListObj := []string{}
	// Load creds
	config, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(awsProfile))
	checkError(err, "load config ")
	s3Client := s3.NewFromConfig(config)

	reqestCount := 0
	// ListObjectsV2 can return only 1000 objects each time. Store the last Key for next call
	var lastKey *string
	for {
		if reqestCount == 0 {
			// the first time don't give "startAfter"
			list_obj, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
			checkError(err, "list s3 obj 0")

			// loading bar low cost
			fmt.Printf(">")
			// foreach object in the list_obj put in slice of of object key name
			for _, object := range list_obj.Contents {
				sliceListObj = append(sliceListObj, *object.Key)
			}

			// if the ListObjectsV2 is not trunct because less 1000 object in last reqest, break the loop
			if list_obj.IsTruncated == false {
				break
			}
			lastKey = list_obj.Contents[len(list_obj.Contents)-1].Key

		} else {
			list_obj, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{Bucket: aws.String(bucketName), StartAfter: lastKey})
			checkError(err, "list s3 obj 1+")

			// loading bar low cost
			fmt.Printf(">")

			for _, object := range list_obj.Contents {
				sliceListObj = append(sliceListObj, *object.Key)
			}

			// if the ListObjectsV2 is not trunct because less 1000 object in last reqest, break the loop
			if list_obj.IsTruncated == false {
				break
			}
			lastKey = list_obj.Contents[len(list_obj.Contents)-1].Key
		}
		reqestCount++
	} // END for
	return sliceListObj
} // END func listFileInBucketS3

func compareObject(listSource []string, listDestination []string) []string {

	listDiff := []string{}

	for _, objectSource := range listSource {
		countainObject := 0
		for _, objectDestination := range listDestination {

			if objectSource == objectDestination {
				countainObject++
			}
		}
		if countainObject == 0 {
			listDiff = append(listDiff, objectSource)
		}
	} // END for objectSource
	return listDiff
} // END func

func downloadListOfS3Object(bucketName string, listObjects []string, pathFolderWhereDownload string, awsProfile string) (err error) {

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(awsProfile))
	checkError(err, "Load config ")
	s3Client := s3.NewFromConfig(config)

	// Download the S3 object using the S3 manager object downloader
	downloader := manager.NewDownloader(s3Client)

	for _, object := range listObjects {

		// Create a file to download to
		file, err := os.Create(pathFolderWhereDownload + "/" + object)
		checkError(err, "Create file local ")
		defer file.Close()

		// download file
		_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(object),
		})
		checkError(err, "When download file :"+object)
	}
	return
}

func uploadListOfFileToS3(bucketName string, listObjects []string, awsProfile string, folderPath string) (err error) {

	fmt.Println("Number of file to upload : ", len(listObjects))
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(awsProfile))
	checkError(err, "load config ")
	s3Client := s3.NewFromConfig(config)

	for _, object := range listObjects {

		file, err := os.Open(folderPath + "/" + object)
		checkError(err, "load local file ")
		defer file.Close()

		uploader := manager.NewUploader(s3Client)
		_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(object),
			Body:   file,
		})
		checkError(err, "error when upload file : "+object)
		// low cost loading bar
		//fmt.Printf("^")
	} // END for

	if len(listObjects) != 0 {
		fmt.Println("\nUpload finish")
	}

	return
}

func cleanLocalFolder(listObjects []string) (err error) {

	for _, object := range listObjects {
		err := os.Remove(object)
		checkError(err, "delete local file name :"+object)
	}
	if len(listObjects) != 0 {
		fmt.Println("\nClean local folder finish")
	}
	return
}

func listBucketS3(awsProfile string) {

	// Load creds
	config, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(awsProfile))

	s3Client := s3.NewFromConfig(config)
	result, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	checkError(err, "list bucket")

	if len(result.Buckets) == 0 {
		fmt.Println("You don't have any buckets!")
	} else {
		for _, bucket := range result.Buckets {
			fmt.Printf("\t%v\n", *bucket.Name)
		}
	}
}

func encryptFile(filename string, folderPathFile string, destinationFolder string) {
	// Reading plaintext file
	plainText, err := ioutil.ReadFile(folderPathFile + "/" + filename)
	checkError(err, "func encryptFile : read file err")

	// Reading key
	key, err := ioutil.ReadFile(keyPath)
	checkError(err, "func encryptFile : read key file err")

	// Creating block of algorithm
	block, err := aes.NewCipher(key)
	checkError(err, "func encryptFile : cipher err")

	// Creating GCM mode
	gcm, err := cipher.NewGCM(block)
	checkError(err, "func encryptFile : cipher GCM err")

	// Generating random nonce
	nonce := make([]byte, gcm.NonceSize())
	checkError(err, "func encryptFile : nonce  err")

	// Decrypt file
	cipherText := gcm.Seal(nonce, nonce, plainText, nil)

	// Writing ciphertext file
	err = ioutil.WriteFile(destinationFolder+"/"+filename, cipherText, 0777)
	checkError(err, "func encryptFile : write file err")

}

func decryptFile(fileEncrypted string, pathFileEncrypted string, pathFileDecrypted string) {
	//func decryptFile(fileEncrypted string, folderPathWhereDownlaod string, folderPathWhereDecryptFile string) {

	// Reading ciphertext file
	cipherText, err := ioutil.ReadFile(pathFileEncrypted + "/" + fileEncrypted)
	checkError(err, "func decryptFile : read file Encrypted")

	// Reading key
	key, err := ioutil.ReadFile(keyPath)
	checkError(err, "func decryptFile : read key file")

	// Creating block of algorithm
	block, err := aes.NewCipher(key)
	checkError(err, "func decryptFile : cipher err")

	// Creating GCM mode
	gcm, err := cipher.NewGCM(block)
	checkError(err, "func decryptFile : cipher GCM err")

	// Deattached nonce and decrypt
	nonce := cipherText[:gcm.NonceSize()]
	cipherText = cipherText[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	checkError(err, "func decryptFile : decrypt file err")

	// Writing decryption content
	err = ioutil.WriteFile(pathFileDecrypted+"/"+fileEncrypted, plainText, 0777)
	checkError(err, "func decryptFile : write file err")

}

func listLocalFile(path string) []string {
	files, err := ioutil.ReadDir(path)
	checkError(err, "func listLocalFile : ReadDir")

	listFiles := []string{}
	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())
		listFiles = append(listFiles, file.Name())

	}
	return listFiles
}

func encryptFilesAndUploadToS3(listFile []string) {

	for _, file := range listFile {
		encryptFile(file, pathSync, pathWorkdir)
	}
	uploadListOfFileToS3(awsBucket, listFile, awsProfile, pathWorkdir)

	fmt.Println("Upload finish")

}

func downloadToLocalAndDecryptFiles(listFileToDownload []string) {

	downloadListOfS3Object(awsBucket, listFileToDownload, pathWorkdir, awsProfile)
	for _, file := range listFileToDownload {
		decryptFile(file, pathWorkdir, pathSync)
	}

	fmt.Println("Download finish")
}

/////////////////////

var (
	keyPath     string = "config/key.txt"
	pathSync    string = "syncFolder"
	pathWorkdir string = "workdir"
	awsProfile  string = "AWS_PERSO"
	awsBucket   string = "florian-drive"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////// MAIN ///////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	//listBucketS3(awsProfile)
	listLocalFile(pathSync)

	listFileBucket := listFileInBucketS3(awsBucket, awsProfile)
	fmt.Println("\nS3 : file count : ", len((listFileBucket)))

	listFileLocal := listLocalFile(pathSync)
	fmt.Println("\nLocal : file count : ", len((listFileLocal)))

	listDiffMissingS3 := compareObject(listFileLocal, listFileBucket)
	fmt.Println("\nMissing files on S3 : ", len(listDiffMissingS3))

	listDiffMissingLocal := compareObject(listFileBucket, listFileLocal)
	fmt.Println("\nMissing files on local : ", len(listDiffMissingLocal))

	encryptFilesAndUploadToS3(listDiffMissingS3)
	downloadToLocalAndDecryptFiles(listDiffMissingLocal)

} // END main
