import logging
import multiprocessing
from common.communication import *
from common.error import MessageReceptionError
from common.utils import *


class SessionHandler:
    def __init__(self, client_socket, number_of_clients, agencies_counter, file_lock, should_exit):
        self.__communication_handler = CommunicationHandler(client_socket)
        self.winners = []
        self.number_of_clients = number_of_clients
        self.agencies_counter = agencies_counter
        self.file_lock = file_lock
        self.should_exit = should_exit

    def start(self):
        """
        Dummy Server loop

        Server that accept new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while True:
            try:
                if self._should_exit() is True:
                    self._graceful_shutdown()
                # TODO: consider creating a Message class
                message_type, payload = self.__communication_handler.recv_msg()
                if message_type == BET_BATCH_MSG_TYPE:
                    batch = payload
                    self._handle_batch_message(batch)

                elif message_type == FINALIZATION_MSG_TYPE:
                    agency_id = payload
                    self._set_agency_as_finished()

                elif message_type == GET_WINNERS_MSG_TYPE:
                    agency_id = payload
                    if self._all_agencies_finished():
                        self._load_winners(agency_id)
                        self.__communication_handler.send_winners(self.winners)
                        break
                    else:
                        self.__communication_handler.send_no_winners_yet()
                else:
                    # TODO: raise an error
                    logging.info(f"Invalid message_type = {message_type}")
            except Exception as e:
                logging.error(f"failed to handle client connection. Error: {e}")
                
        self.__communication_handler.close_current_connection()

    def _handle_batch_message(self, batch: list[Bet]):
        try:
            bets = batch
            with self.file_lock:
                store_bets(bets)
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            self.__communication_handler.send_batch_success()
        except Exception as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            self.__communication_handler.send_batch_failure()

    def _all_agencies_finished(self) -> bool:
        return self.agencies_counter.value == self.number_of_clients 

    def _set_agency_as_finished(self):
        self.agencies_counter.value += 1
        if self._all_agencies_finished:
            logging.info(f"action: sorteo | result: success")

    def _load_winners(self, agency_id):
        bets = load_bets()
        for bet in bets:
            if has_won(bet) and bet.agency == agency_id:
                self.winners.append(bet.document)

    def _should_exit(self) -> bool:
        return self.should_exit.value == 1
    
    def _graceful_shutdown(self):
        # TODO: check if there is anything else to close
        self.__communication_handler.close_current_connection()
        exit(0)
