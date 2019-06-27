# gphotos-cli: A command line interface for Google Photos

A command line interface app for Google photos for doing bulk tasks such as mass filtered downloads
and mass uploads.

## Features

* Parallelized uploading and downloading
* Filtering by date and category type for downloads
* Album creation
* Uploads and downloads to/from albums (can only upload to albums created by this app due to Google's API limitations)
* Exponential backoff for large uploads and downloads, I've personally tested jobs 100 GB+ large
* Handles all raw video and image types supported by Google photos

# Installation

With Go already installed and setup, just call:

`go get -u github.com/trinhdrew1418/gphotos-cli`

Before usage, you first need to grant the application account authorization:

`gphotos-cli init`

# Commands
### init

This initializes the authorization token needed to access your account

`gphotos-cli init`

#### options

* `-c` or `--credential-path`
Designate the path of the credenntial file you'd alternatively like to use.

### push

Command to upload files to your google photos library.

`gphotos-cli push [-OPTIONS] [FILE 1] [FOLDER 1] ...`

#### options

* `-r` or `--recursive`
    Recursively traverses folders in arguments for file uploads

* `-s` or `--select`
    Pulls up a menu to select available albums from to upload to. NOTE: you can ONLY upload to
    albums that you've created through this app. This is a limitation of the google photos api.

* `-v` or `--verbose`
    This lists out all of the files you'll be uploading. Useful if you want to know which files have
    valid extensions.
 
* `-a` or `--album`
    Input the name of the album you want to upload to

### pull

Command to download files

`gphotos-cli pull [-OPTIONS]`

Follow the prompts to select your filters. It will download your files into folders separated by year and month. Each
file will be named with its day of creation and time.

You can filter up to 10 of the following categories:

* animals
* landmarks
* pets
* utility
* birthdays
* landscapes
* receipts
* weddings
* cityscapes
* night
* screenshots
* whiteboards
* arts
* crafts
* fashion
* documents
* people
* selfies
* houses
* gardens
* flowers
* holidays
* travel
* food
* performances
* sport

#### options

* `-d [PATH]`
    define the directory path you want to download your files to.
    
* `-s` or `--select`
    pull up a selection menu of albums to download instead. The files will be downloading
    into a folder of the name of the album. The files will not be organized further beyond
    this. As a limitation of the API, filtering in conjunction with an album request is 
    not possible.
    
* `-a` or `--album`
    Input the name of the album you want to download from.

### `albums`

Create albums by calling

`gphotos-cli albums create [TITLE OF ALBUM]`

More subcommands coming later.

# Caveats

* You may unexpectedly hit a quota limit due to the application coming with some default credentials.
* Due to a google photos api limitation, you can only upload to albums created by the app
* When downloading files, its undetermined how many photos in total you`ll be downloading. The API uses pagination for large
requests and instead of waiting for possibly several page requests for the full download request, it is faster to
concurrently do requests for pages whilst downloading what's available. The current default page size is 25.
* The API does not support the retrieval of raw files. If you try to pull a raw file, it will return a compressed jpeg. Downloading
raw files still works through the website though.

# Credits
Some of these packages were a huge help; either directly, as reference code, or as inspiration.

* [cobra](https://github.com/spf13/cobra)
* [gphotos-uploader-cli](https://github.com/nmrshll/gphotos-uploader-cli)
* [mpb](https://github.com/vbauerster/mpb)
* [promptui](https://github.com/manifoldco/promptui)
* [drive](https://github.com/odeke-em/drive)


## TODOs

* Mass moving existing photos to albums
* Search by metadata: specific filetypes, camera types, etc.
* Resumable uploads and data chunking for large file uploads
