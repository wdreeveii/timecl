
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>

#include "usart.h"
#include "gb_client.h"
#include "iocontrol.h"

uint16_t crc16_update(uint16_t crc, uint8_t a)
{
	int i;

	crc ^= a;
	for (i = 0; i < 8; i++)
	{
		if (crc & 1)
			crc = (crc >> 1) ^ 0xA001;
		else
			crc = (crc >> 1);
	}

	return crc;
}

void print_packet(char *buffer, int length) {
	struct mheader_t *msg_header = (struct mheader_t*)(buffer + 1);
	printf("pkt:\n");
	printf("destination: %d\n", msg_header->destination);
	printf("type: %d\n", msg_header->mtype);
	printf("length: %d\n", msg_header->length);
	printf("mac: %ld\n", msg_header->mac);
	/*printf("crc: %d\n", msg_header->crc);
	printf("crc hex: %#x\n", msg_header->crc);
	for (int i = sizeof(struct mheader_t) +2; i < length; i++)
		printf("%c", *(buffer + i));
	printf("\n");
	for (int i = sizeof(struct mheader_t) +2; i < length; i++)
		printf("%#x ", *(buffer + i));*/
	printf("\n");
	//printf("payload crc: %#x\n", *((uint16_t*)(buffer + (length - 2))) );
}

uint16_t header_crc(char * buffer)
{
	uint16_t checksum = 0xffff;
	for (int i = 0; i < sizeof(struct mheader_t) - 1; i++)
		checksum = crc16_update(checksum, buffer[i]);
	return checksum;
}
void send_packet(struct message_t * msg) {
	char data[sizeof(struct mheader_t) + 2 + msg->payload_len + 2];
	struct mheader_t * header;
	uint16_t checksum = 0xffff;
	data[0] = 'A';
	data[sizeof(struct mheader_t) + 1] = 'A';
	
	memcpy((data + 1), &(msg->header), sizeof(struct mheader_t));
	memcpy((data + sizeof(struct mheader_t) + 2), msg->payload, msg->payload_len);
	
	header = ((struct mheader_t *)(data + 1));
	header->length = msg->payload_len + 2;
	header->crc = header_crc(data);
	for (int i = 0; i < sizeof(struct mheader_t) + 2 + msg->payload_len; i++)
		checksum = crc16_update(checksum, data[i]);
	
	*((uint16_t *)(data + sizeof(struct mheader_t) + 2 + msg->payload_len)) = checksum;
	USART_Send(0, (uint8_t *)data, sizeof(struct mheader_t) + 2 + msg->payload_len + 2);
}

void mk_interrogate_reply(struct message_t * msg) {
	uint8_t num_ports = io_num_ports();
	uint8_t payload_len = (num_ports*2)+1+6+11;
	char payload[payload_len];
	payload[0] = 5;
	strcpy(payload+1, "DEV01");
	payload[1 + 5] = 10;
	strcpy(payload+1+5+1, "1234567890");
	payload[1 + 5 + 1 + 10] = num_ports;
	for(int i = 0; i < num_ports; i++) {
		payload[1+5+1+10+1+(i*2)] = io_get_type(i);
		payload[1+5+1+10+1+(i*2)+1] = i;
	}
	msg->header.mtype = INTERROGATE_REPLY;
	msg->payload = payload;
	msg->payload_len = payload_len;
	send_packet(msg);
}

void send_value_reply(struct message_t * msg, uint8_t message_type) {
	uint8_t num_ports = io_num_ports();
	char payload[1+(3*num_ports)];
	payload[0] = 0;
	for (int i = 0; i < num_ports; i++) {
		uint8_t type = io_get_type(i);
		if (type == PORT_BINPUT) {

			payload[1+(3*payload[0])] = i;
			payload[2+(3*payload[0])] = io_read(i);
			payload[3+(3*payload[0])] = 0;
			payload[0]++;
		} else
		if (type == PORT_AINPUT) {
			payload[1+(3*payload[0])] = i;
			uint16_t val = io_aread(i);
			payload[2+(3*payload[0])] = (uint8_t)val;
			payload[3+(3*payload[0])] = (uint8_t)(val>>8);
			payload[0]++;
		}
	}

	msg->header.mtype = message_type;
	msg->payload = payload;
	msg->payload_len = 1 + payload[0]*3;
	send_packet(msg);
}
	
void do_set(struct message_t * msg, char * payload, int length) {
	if (length > 0) {
		uint8_t num_ports = *payload;
		if (length >= 1 + (num_ports * 3)) {
			for (int i = 0; i < num_ports; i++) {
				uint8_t port = *(payload + 1 + (i*3));
				uint16_t value = *( (uint16_t *)( payload + 2 + (i*3) ));
				uint8_t val = value;
				iocontrol(port, val);
				printf("set: %d : %d\n", port, val);
			}
		}
	}
	send_value_reply(msg, SET_REPLY);
}

void process_packet(char *buffer, int length) {
	struct mheader_t *header;
	struct message_t msg;
	unsigned long tmp;
	static uint8_t addr = 0;
	static uint32_t mac = 128;
	
	//print_packet(buffer, length);
	header = (struct mheader_t *)(buffer + 1);
	
	if (addr == 0) {
		if (header->destination == 0) {
			if (header->mtype == FIND_DEVICES && header->mac == 0) {
				// some random delay
				msg.header.destination = 1;
				msg.header.mac = mac;
				msg.header.mtype = NEED_ADDR;
				msg.payload = "NEED IP";
				msg.payload_len = 7;
				send_packet(&msg);
			}
			if (header->mtype == ACK_DEVICE && header->mac == mac) {
				msg.header.destination = 1;
				msg.header.mac = mac;
				msg.header.mtype = ACK_REPLY;
				msg.payload = "ACK";
				msg.payload_len = 3;
				send_packet(&msg);
				tmp = strtoul(buffer + 2 + sizeof(struct mheader_t), NULL, 10);
				addr = 0xFF&tmp;
				printf("THE ADDR: %d\n", addr);
			}
		}
	}
	else {
		if (header->destination == addr && header->mac == mac) {
			msg.header.destination = 1;
			msg.header.mac = 128;
			
			if (header->mtype == SET) {
				return do_set(&msg, 
							buffer + 2 + sizeof(struct mheader_t),
							length-(2+2+sizeof(struct mheader_t)));
			}
			else if (header->mtype == PING) {

				msg.header.mtype = PING_REPLY;
				msg.payload = "HAHAHA";
				msg.payload_len = 6;
			}
			else if (header->mtype == INTERROGATE)
			{
				return mk_interrogate_reply(&msg);
			}
			else {
				msg.header.mtype = UNKNOWN_REPLY;
				msg.payload = "Unknown Cmd";
				msg.payload_len = 11;
			}
			send_packet(&msg);
		}
		else if (header->destination == 0 && header->mtype == RESET_NETWORK && header->mac == 0) {
			// received a reset address request. 
			addr = 0;
		}
	}
	
}

void gb_push_char(char *buffer, char in) {
	struct mheader_t* msg_header = (struct mheader_t*)(buffer + 1);
	static int end = 0;
	uint16_t checksum = 0xffff;
	if (end == 500) {
		end = 0;
	}
	buffer[end] = in;
	end++;
	if (end == sizeof(struct mheader_t) + 2) {
		if (buffer[0] != 'A' || buffer[sizeof(struct mheader_t) + 1] != 'A') {
			goto remove_first;
		}
		for (int i = 0; i < sizeof(struct mheader_t) - 1; i++)
			checksum = crc16_update(checksum, buffer[i]);
			
		if (checksum != msg_header->crc) {
			goto remove_first;
		}
	}
	if (end > sizeof(struct mheader_t) + 2) {
		if (end - (sizeof(struct mheader_t) + 2) == msg_header->length) {
			checksum = 0xffff;
			for (int i = 0; i < end - 2; i++)
				checksum = crc16_update(checksum, buffer[i]);
			uint16_t * crc = (uint16_t*)(buffer + (end - 2));
			if (checksum == *crc)
			{
				process_packet(buffer, end);
			}
			end = 0;
		}
	}
	
	return;
	
remove_first:
	memmove(buffer, buffer+1, end);
	end--;
	return;
}
