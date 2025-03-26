import logging
from multiprocessing import Process, Lock
from multiprocessing.sharedctypes import Value
from common.session_handler import SessionHandler
from common.communication import *
from common.error import MessageReceptionError
from common.utils import *


class Server:
    def __init__(self, port, listen_backlog, number_of_clients):
        # self.__communication_handler = CommunicationHandler(port, listen_backlog)
        # self.finished_agencies = set()
        # self.winners_by_agency = {}
        self.number_of_clients = int(number_of_clients)
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Set socket timeout to check for the signal flag
        self._server_socket.settimeout(SOCKET_TIMEOUT)


    def run(self):
        """
        Dummy Server loop

        Server that accept new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        # This counter holds the number of agencies that have finalized sending their bets
        agencies_counter = Value('i', 0)
        processes = []
        file_lock = Lock()
        # TODO: replace the while loop with a for loop that loops for number_of_clients
        while True:
            # TODO: find a way for the server to know when to join the processes
            client_socket = self.accept_new_connection()
            session_handler = SessionHandler(client_socket, self.number_of_clients, agencies_counter, file_lock)
            p = Process(target=session_handler.start)
            processes.append(p)
            p.start()

    def accept_new_connection(self):
        """
        Accept new connections

        Function blocks for SOCKET_TIMEOUT seconds until a connection to a client is made.
        Then connection created is printed and returned.
        If a SIGTERM signal has been received by the time the socket times out,
        then the server exits gracefully.
        """

        for _ in range(self.number_of_clients):
            # TODO: add graceful shutdown
            # if self._received_sig_term:
            #     self.__exit_gracefully()
            try:
                # Connection arrived
                logging.info('action: accept_connections | result: in_progress')
                c, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                return c
            except socket.timeout:
                # This timeout allows to check for the SIGTERM signal more regularly
                continue

    # def _handle_batch_message(self, batch: list[Bet]):
    #     try:
    #         bets = batch
    #         store_bets(bets)
    #         logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
    #         self.__communication_handler.send_batch_success()
    #     except Exception as e:
    #         logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
    #         self.__communication_handler.send_batch_failure()

    # def _all_agencies_finished(self) -> bool:
    #     agencies_ids = list(range(1, self.number_of_clients + 1))
    #     for id in agencies_ids:
    #         if id not in self.finished_agencies:
    #             return False
    #     return True

    # def _set_agency_as_finished(self, agency: int):
    #     self.finished_agencies.add(agency)

    # def _load_winners(self):
    #     bets = load_bets()
    #     for bet in bets:
    #         if has_won(bet):
    #             self.winners_by_agency.setdefault(bet.agency, []).append(bet.document)
