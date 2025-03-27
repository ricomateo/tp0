import socket
import signal
import logging
import sys
from common.utils import Bet
from common.error import MessageReceptionError

SOCKET_TIMEOUT = 5
BET_BATCH_MSG_TYPE = 0
BATCH_CONFIRMATION_MSG_TYPE = 1

BATCH_FAILURE_STATUS = 0
BATCH_SUCCESS_STATUS = 1

class CommunicationHandler:
    def __init__(self, port, listen_backlog):
        # Set the SIGTERM handler
        signal.signal(signal.SIGTERM, self.__sigterm_handler)
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Set socket timeout to check for the signal flag
        self._server_socket.settimeout(SOCKET_TIMEOUT)
        self._client_sock = None
        self._received_sig_term = False

    def accept_new_connection(self):
        """
        Accept new connections

        Function blocks for SOCKET_TIMEOUT seconds until a connection to a client is made.
        Then connection created is printed and returned.
        If a SIGTERM signal has been received by the time the socket times out,
        then the server exits gracefully.
        """

        while True:
            if self._received_sig_term:
                self.__exit_gracefully()
            try:
                # Connection arrived
                logging.info('action: accept_connections | result: in_progress')
                c, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                self._client_sock = c
                return
            except socket.timeout:
                # This timeout allows to check for the SIGTERM signal more regularly
                continue

    def recv_msg(self):
        """
        Reads a message from the current client socket
        """
        try:
            message_type = int.from_bytes(self._recv_exact(1), "big")
            if message_type == BET_BATCH_MSG_TYPE:   
                bets = self.__decode_bet_batch()
                return bets
            else:
                raise MessageReceptionError("Invalid message type")
        
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            raise MessageReceptionError(e)
        except Exception as e:
            raise MessageReceptionError(e)

    def send_batch_success(self):
        self.__send_batch_status(BATCH_SUCCESS_STATUS)

    def send_batch_failure(self):
        self.__send_batch_status(BATCH_FAILURE_STATUS)

    def close_current_connection(self):
        self._client_sock.close()

    def __decode_bet_batch(self) -> list[Bet]:
        batch_size = int.from_bytes(self._recv_exact(4), "big")
        bets = []
        for _ in range(batch_size):
            bet = self.__decode_bet_info()
            bets.append(bet)
        return bets
    
    def __send_batch_status(self, status: int):
        message_type = BATCH_CONFIRMATION_MSG_TYPE
        # Send message type
        self._client_sock.sendall(message_type.to_bytes(1, "big"))
        # Send status
        self._client_sock.sendall(status.to_bytes(1, "big"))

    def __decode_bet_info(self) -> Bet:
        """
        Reads and returns a Bet message from the current socket connection.
        """
        # Deserialize the fields
        agency = self.__recv_str()
        name = self.__recv_str()
        last_name = self.__recv_str()
        document = self.__recv_str()
        birthdate = self.__recv_str()
        number = self.__recv_str()
        
        return Bet(agency, name, last_name, str(document), birthdate, str(number))
    
    def __recv_str(self) -> str:
        """
        Reads and return a string from the current socket connection.
        The string is decoded by first reading its length, and then the value.
        """
        str_len = int.from_bytes(self._recv_exact(1), "big")
        string = str(self._recv_exact(str_len), 'utf-8')
        return string

    def __sigterm_handler(self, signum, _):
        if signum == signal.SIGTERM:
            # This assignment is atomic
            self._received_sig_term = True

    def __exit_gracefully(self):
        logging.info("Exiting gracefully")
        logging.info("Shutting down server")
        self._server_socket.shutdown(socket.SHUT_RDWR)
        if self._client_sock is not None:
            logging.info("Closing socket connection")
            self._client_sock.close()
        sys.exit(0)

    def _recv_exact(self, n):
        """
        Reads exactly n bytes from the socket, and returns the data.
        If the connection is closed, raises an exception.
        """
        data = bytes()
        while len(data) < n:
            received_bytes = self._client_sock.recv(n - len(data))
            if not received_bytes:
                raise ConnectionError("Connection closed")
            data += received_bytes
        return data
