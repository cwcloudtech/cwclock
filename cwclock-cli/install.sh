parse_commandline() {
  while test $# -gt 0
  do
    key="$1"
	case "$key" in
	  -i|--install-dir)
	    PARSED_INSTALL_DIR="$2"
		shift
	   ;;
	  -b|--bin-dir)
	    PARSED_BIN_DIR="$2"
		shift
	   ;;
	  -u|--update)
	    PARSED_UPGRADE="yes"
	  ;;
	  *)
	   die "Got an unexpected argument: $1"
	  ;;
    esac
	shift
  done
}

set_global_vars() {
  ROOT_INSTALL_DIR=${PARSED_INSTALL_DIR:-/usr/local/cwclock}
  BIN_DIR=${PARSED_BIN_DIR:-/usr/local/bin}
  UPGRADE=${PARSED_UPGRADE:-no}
  EXE_NAME="cwclock"
  INSTALLER_DIR="$( cd "$( dirname "$0" )" >/dev/null 2>&1 && pwd )"
  INSTALLER_EXE="$INSTALLER_DIR/$EXE_NAME"
  CWCLOCK_EXE_VERSION=$($INSTALLER_EXE --version | cut -d ' ' -f 1 | cut -d '/' -f 2 )
  INSTALL_DIR="$ROOT_INSTALL_DIR/$CWCLOCK_EXE_VERSION"
  CURRENT_INSTALL_DIR="$ROOT_INSTALL_DIR"
  CURRENT_CWCLOCK_EXE="$CURRENT_INSTALL_DIR/$EXE_NAME"
  BIN_CWCLOCK_EXE="$BIN_DIR/$EXE_NAME"
}

create_install_dir() {

  mkdir -p "$INSTALL_DIR" || exit 1
  {
    setup_install &&
    create_current_symlink
  } 
}

check_preexisting_install() {

  if [ -d "$CURRENT_INSTALL_DIR" ] && [ "$UPGRADE" = "no" ]
  then
    die "Found preexisting CWClock CLI installation: $CURRENT_INSTALL_DIR. Please rerun install script with --update flag."
  fi
  if [ -d "$INSTALL_DIR" ]
  then
    echo "Found same CWClock CLI version: $INSTALL_DIR. Skipping install."
    exit 0
  fi
}

setup_install() {
  cp -r "$INSTALLER_EXE" "$INSTALL_DIR"
}



create_current_symlink() {
  ln -snf "$INSTALL_DIR/$EXE_NAME" "$CURRENT_INSTALL_DIR"
}

create_bin_symlinks() {
  mkdir -p "$BIN_DIR"
  ln -sf "$CURRENT_CWCLOCK_EXE" "$BIN_CWCLOCK_EXE"
}

die() {
	err_msg="$1"
	echo "$err_msg" >&2
	exit 1
}

main() {
  parse_commandline "$@"
  set_global_vars
  check_preexisting_install
  create_install_dir
  create_bin_symlinks
  echo "You can now run: $BIN_CWCLOCK_EXE --version"
  exit 0
}

main "$@" || exit 1