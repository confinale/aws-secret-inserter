# secrets inserter

Allows you to replace a secret in a file using secrets manager. 
`::SECRET:secret-name:SECRET::` will be replaced with your `secret-name` stored in AWS secrets manager.

You can either replace it inline `secrets-insert -f myFile.txt -i` or just outputting it to stdout `secrets-insert -f myFile.txt`. If you pass `-fail` the execution will stop if any matched secret could not be replaced. 