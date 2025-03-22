# Build the netcat image, capturing its output
BUILD_OUTPUT=$(docker build . -q -f netcat.Dockerfile -t nc)

MESSAGE="hello"
HOST=server
PORT=12345
NETWORK=tp0_testing_net

# Start the container, send "hello" to the server, and save the response in 'output.txt'
RESPONSE=$(docker run --rm --network=$NETWORK --name=netcat nc -c "echo $MESSAGE | nc $HOST $PORT" > output.txt 2> error.txt)

RESPONSE=$(cat output.txt | tr -d '\r')
rm output.txt
rm error.txt

if [[ "$RESPONSE" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

