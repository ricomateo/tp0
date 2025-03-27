import logging
from multiprocessing import Process, Lock, Barrier
from multiprocessing.sharedctypes import Value
from common.session_handler import SessionHandler
from common.communication import *
from common.error import MessageReceptionError
from common.utils import *

BARRIER_TIMEOUT = 2

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

        # Shared flag that is set to 1 when the server receives a SIGTERM signal.
        self.should_exit = Value('i', 0)
        self.number_of_clients = int(number_of_clients)
        self.sessions = []
        self.barrier = Barrier(self.number_of_clients, timeout=BARRIER_TIMEOUT)

    def run(self):
        """
        The Server listens for client connections and spawns a SessionHandler process to handle the session.

        Each of the sessions share some shared variables:
            * barrier: allows to synchronize the sessions to send the winners at the same time.
            * file_lock: a lock to synchronize the access to the file.
            * should_exit: a flag that is set to 1 when the server receives a SIGTERM signal. The session handlers constantly
            check this value to know if they should exit.
        """
        file_lock = Lock()
        for _ in range(self.number_of_clients):
            if self._should_exit() is True:
                break
            client_socket = self.accept_new_connection()
            session_handler = SessionHandler(client_socket, self.number_of_clients, file_lock, self.should_exit, self.barrier)
            session = Process(target=session_handler.start)
            self.sessions.append(session)
            session.start()
        
        self._graceful_shutdown()


    def accept_new_connection(self):
        """
        Blocks for SOCKET_TIMEOUT waiting for connections. The timeout is required to check for the
        'should_exit' flag.
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
        """
        SIGTERM handler. Sets the shared flag 'should_exit' to 1.
        """
        if signum == signal.SIGTERM:
            self.should_exit.value = 1

    def _graceful_shutdown(self):
        """
        Shuts down the server gracefully, closing the socket, and joining all the sessions.
        Each session closes its own socket.
        """
        logging.info("Shutting down")
        logging.info("Closing server socket")
        self._server_socket.close()
        logging.info("Joining sessions")
        for session in self.sessions:
            session.join()
        sys.exit(0)


    def _should_exit(self):
        return self.should_exit.value == 1
    