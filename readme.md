## Usage

ssh server iostat -x 4 | ./iostat-parser -H server -i http://influxdb:8083/write?db=iostat


