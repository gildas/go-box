# go-box

![GoVersion](https://img.shields.io/github/go-mod/go-version/gildas/go-box)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/gildas/go-box)
[![License](https://img.shields.io/github/license/gildas/go-box)](https://github.com/gildas/go-box/blob/master/LICENSE)
[![Report](https://goreportcard.com/badge/github.com/gildas/go-box)](https://goreportcard.com/report/github.com/gildas/go-box)  

go-box is a Go client library for accessing the [Box.com API](https://developer.box.com/).

## Installation

```bash
go get github.com/gildas/go-box
```

## Usage

After importing the package in your code:

```go
import "github.com/gildas/go-box"
```

You can create a new client with:

```go
client := box.NewClient(context)
```

### Authentication

To let the client authenticate with [nox.com](https://box.com), you need to provide credentials and authenticate:

```go
client := box.NewClient(context)
creds := box.Credentials{
	ClientID:     "your-client-id",
	ClientSecret: "your-client-secret",
	EnterpriseID: "your-enterprise-id",
	AppAuth: box.AppAuth{
		PublicKeyID: "your-public-key-id",
		PrivateKey:  "your-private-key",
		PassPhrase: "your-passphrase",
	},
}
if err := client.Authenticate(creds); err != nil {
	log.Errorf("Failed to authenticate.", err)
}
```

### Uploading a file

To upload a file you need to know the owner Id of the folder you want to upload the file to:

```go
import "github.com/gildas/go-box"
import "github.com/gildas/go-logger"
import "github.com/gildas/go-request"

func UploadFile(context context.Context, client *box.Client, ownerID string, filename string, reader io.Reader) (*box.FileCollection, error) {
	log := logger.Must(logger.FromContext(context)).Child("box", "upload", "owner", ownerID)

	// Get the folder for the owner
	log.Debugf("Getting the folder for the owner.")
	folder, err := client.Folders.FindByName(context, ownerID)
	if err != nil {
		log.Errorf("Failed to get the folder for the owner.", err)
		return nil, err
	}

	return client.Files.Upload(context, &box.UploadOptions{
		Filename: filename,
		Parent:   folder.AsPathEntry(),
		Content:  request.ContentFromReader(reader),
	})
}
```

You can also upload payloads directly:

```go
files, err := client.Files.Upload(context, &box.UploadOptions{
	Parent:   folder.AsPathEntry(),
	Filename: "test.json",
	Payload: struct {
		Name string `json:"name"`
	}{
		Name: "Test",
	},
})
```

### Finding files

To find files, you can use the `Find` methods:

```go
root, err := client.Folders.FindByID(context, "0")
entry, err := client.Files.FindByName(context, "test.json", root.AsPathEntry())
```

You can also find files by Id (coming from a previous upload, for example):

```go
root, err := client.Folders.FindByID(context, "0")
entry, err := client.Files.FindByID(context, "1234567890", root.AsPathEntry())
```

### Downloading a file

To download a file, you need the entry (see above):

```go
downloaded, err := client.Files.Download(context, root.ItemCollection.Paths[0])
```

`downloaded` is a [request.Content](https://pkg.go.dev/github.com/gildas/go-request#Content) that you can use to read the content of the file.

### Deleting a file

To delete a file:

```go
err = client.Folders.Delete(context, root.ItemCollection.Paths[0])
```

### Creating a folder

To create a folder, you need their parent folder:

```go
root, err := client.Folders.FindByID(context, "0")
folder, err := client.Folders.Create(context, &box.FolderEntry{Name: "New Folder", Parent: root.AsPathEntry()})
```

### Finding a folder

To find a folder, you can use the `Find` methods:

```go
root, err := client.Folders.FindByID(context, "0")
entry, err := client.Folders.FindByName(context, "New Folder", root.AsPathEntry())
```

You can also find folders by Id (coming from a previous creation, for example):

```go
root, err := client.Folders.FindByID(context, "0")
entry, err := client.Folders.FindByID(context, "1234567890", root.AsPathEntry())
```

### Deleting a folder

To delete a folder:

```go
err = client.Folders.Delete(context, root.ItemCollection.Paths[0])
```

### Shared Links

To create a shared link:

```go
link, err := client.SharedLinks.Create(context, &entry, nil)
```

You can also give some options to the shared link:

```go
link, err := client.SharedLinks.Create(context, &entry, &box.SharedLinkOptions{
	Access: "open",
	Permissions: box.SharedLinkPermissions{
		CanDownload: true,
	},
	UnsharedAt: time.Now().Add(24 * time.Hour),
})
```

### Logging

[go-box](https://github.com/gildas/go-box) uses [go-logger](https://github.com/gildas/go-logger) for logging. The logs are [bunyan](https://github.com/trentm/node-bunyan) compatible. You can either use the [bunyan CLI](https://github.com/trentm/node-bunyan?tab=readme-ov-file#cli-usage) to read them or use the [lv](https://github.com/gildas/lv) tool to read them.

If you don't provider a logger, the client will create one for you that logs to `os.Stderr` with the `INFO` level.

To provide your own logger, you can do something like this (please have a look at [go-logger](https://github.com/gildas/go-logger/blob/master/README.md) for more information):

```go
log := logger.Create("myapp", &logger.FileStream{Path: "myapp.log", FilterLevels: logger.NewLevelSet(logger.DEBUG)})
defer log.Flush()

client := box.NewClient(log.ToContext(context))
```
