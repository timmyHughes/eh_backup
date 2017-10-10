# eh_backup
go based ExtraHop configuration and customization backup

NAME:
   backup_eh - A new cli application

USAGE:
   eh_backup [global options] command [command options] [arguments...]
   
VERSION:
   1.2
   
COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --file value, -f value    Configuration File
   --debug, -d               Debug Output
   --server value, -s value  REQUIRED: Extrahop API Host
   --key value, -k value     REQUIRED: Key for accessing Extrahop's REST API
   --type value, -t value    REQUIRED: Backup type: [runningconfig|customizations]
   --node value, -n value    Node Identifier
   --help, -h                show help
   --version, -v             print the version
