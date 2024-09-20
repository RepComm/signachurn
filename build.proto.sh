SRC_DIR="./proto"
DST_DIR="./"
protoc \
-I=$SRC_DIR \
--go_out=$DST_DIR \
$SRC_DIR/scan.proto
