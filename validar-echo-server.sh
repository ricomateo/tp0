# Build the netcat image, capturing its output
BUILD_OUTPUT=$(docker build . -q -f netcat.Dockerfile -t nc)

MESSAGE="hello"
HOST=server
PORT=12345
NETWORK=tp0_testing_net
NAME=netcat

# Start the container, send "hello" to the server
docker run -d --network=$NETWORK --name=$NAME nc -c "echo $MESSAGE | nc $HOST $PORT"

# Get the server response through 'docker logs'
RESPONSE=$(docker logs $NAME)

# Remove the container
docker rm $NAME

# Check the response
if [[ "$RESPONSE" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

