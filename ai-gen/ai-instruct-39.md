# AI instruction 39

## External connections

I want external connections for Google drive and S3/object storage optional for every organizations and stored in data's payload in a `external_connections` list.

For S3 connection store the usual information: endpoint, bucket_name, region, access_key, secret_key and for Google drive a service account (in base64) and a folder id.

Add form with a "+" button selecting the type of external connections than adapt the configuration's form (every field is mandatory).

## Invoice generation or update

On every external connection, create a folder `YYYY/MM` and upload the invoice there.
If a file already exists with the same name, replace it.

## Invoice deletion

When an invoice is deleted, delete it from the external connection too.

## Update existing invoice

On the invoice datatable on the same colum of delete button, add an upload icon that will reuploade/replace the invoice on all the external connections.
