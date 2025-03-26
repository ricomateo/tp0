import logging
from multiprocessing import Process, Lock
from multiprocessing.sharedctypes import Value
from common.session_handler import SessionHandler
from common.communication import *
from common.error import MessageReceptionError
from common.utils import *


class Server:
    def __init__(self, port, listen_backlog, number_of_clients):
        # Set the SIGTERM handler
        signal.signal(signal.SIGTERM, self.__sigterm_handler)

        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Set socket timeout to check for the signal flag
        self._server_socket.settimeout(SOCKET_TIMEOUT)
        self.should_exit = Value('i', 0)
        self.number_of_clients = int(number_of_clients)
        self.sessions = []

    def run(self):
        """
        """
        # This counter holds the number of agencies that have finalized sending their bets
        agencies_counter = Value('i', 0)
        
        file_lock = Lock()
        # TODO: replace the while loop with a for loop that loops for number_of_clients
        for _ in range(self.number_of_clients):
            if self._should_exit() is True:
                self._graceful_shutdown()
            # TODO: find a way for the server to know when to join the processes
            client_socket = self.accept_new_connection()
            session_handler = SessionHandler(client_socket, self.number_of_clients, agencies_counter, file_lock, self.should_exit)
            session = Process(target=session_handler.start)
            self.sessions.append(session)
            session.start()
        
        for session in self.sessions:
            session.join()


    def accept_new_connection(self):
        """
        """

        while True:
            if self._should_exit():
                self._graceful_shutdown()
            try:
                # Connection arrived
                logging.info('action: accept_connections | result: in_progress')
                c, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                return c
            except socket.timeout:
                # This timeout allows to check for the SIGTERM signal more regularly
                continue
    
    def __sigterm_handler(self, signum, _):
        if signum == signal.SIGTERM:
            self.should_exit.value = 1

    def _graceful_shutdown(self):
        # TODO: add more logs!
        logging.info("Shutting down!")
        self._server_socket.close()
        for session in self.sessions:
            session.join()
        sys.exit(0)


    def _should_exit(self):
        logging.info(f"self.should_exit.value = {self.should_exit.value}")
        return self.should_exit.value == 1
    