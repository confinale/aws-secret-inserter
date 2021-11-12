# secrets inserter

Allows you to replace a secret in a file using secrets manager. 
`::SECRET:secret-name:SECRET::` will be replaced with your `secret-name` stored in AWS secrets manager.

If you want to access key value secrets you can do that by specifying the requested key after a pipe: `::SECRET:sample1|key1:SECRET::`.

You can either replace it inline `secrets-insert -f myFile.txt -i` or just outputting it to stdout `secrets-insert -f myFile.txt`. If you pass `-fail` the execution will stop if any matched secret could not be replaced. 