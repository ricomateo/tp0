import logging
from common.communication import *
from common.utils import *


class SessionHandler:
    def __init__(self, client_socket, number_of_clients, file_lock, should_exit_lock, should_exit, barrier):
        # Communication handler, in charge of sending and receiving messages
        self.__communication_handler = CommunicationHandler(client_socket)
        # A list of the winners 
        self.winners = []
        self.number_of_clients = number_of_clients
        # File lock used to synchronize file access
        self.file_lock = file_lock
        # Lock to protect the 'should_exit' flag
        self.should_exit_lock = should_exit_lock
        # Shared flag used to know whether the session handler must exit
        self.should_exit = should_exit
        # Barrier to wait for all the clients to send their bets
        self.barrier = barrier

    def start(self):
        """
        The Session Handler receives messages from the client and handles them.
        Once the bet winners have been sent to the client, the Session handler exits.
        """
        while True:
            try:
                if self._should_exit() is True:
                    break
                message_type, payload = self.__communication_handler.recv_msg()
                if message_type == BET_BATCH_MSG_TYPE:
                    batch = payload
                    self._handle_batch_message(batch)

                elif message_type == FINALIZATION_MSG_TYPE:
                    continue

                elif message_type == GET_WINNERS_MSG_TYPE:
                    agency_id = payload
                    self._handle_get_winners_message(agency_id)
                    break
                else:
                    logging.info(f"invalid message_type = {message_type}")
            except Exception as e:
                logging.error(f"failed to handle client connection. Error: {e}")
                break
                
        self.__communication_handler.close_current_connection()

    def _handle_get_winners_message(self, agency_id):
        """
        Handles the get_winners message.
        It waits on the barrier to synchronize with the other Session Handlers.
        Once all the agencies have sent their bets, the Session handlers respond
        with the winners to each of the agencies.
        """
        # All the session handlers are synchronized here 
        self.barrier.wait()
        logging.info(f"action: sorteo | result: success")
        self._load_winners(agency_id)
        self.__communication_handler.send_winners(self.winners)

    def _handle_batch_message(self, batch : list[Bet]):
        """
        Handles the batch message, storing the bets and sending a response (success or failure).
        """
        try:
            bets = batch
            with self.file_lock:
                store_bets(bets)
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            self.__communication_handler.send_batch_success()
        except Exception as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            self.__communication_handler.send_batch_failure()

    def _load_winners(self, agency_id):
        """
        Loads the winners of the given agency.
        """
        bets = load_bets()
        for bet in bets:
            if has_won(bet) and bet.agency == agency_id:
                self.winners.append(bet.document)

    def _should_exit(self) -> bool:
        with self.should_exit_lock:
            return self.should_exit.value == 1
