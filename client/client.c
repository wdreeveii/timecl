
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <termios.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>

#define MAX_PAYLOAD_LEN 250

struct mheader_t {
	uint8_t destination;
	uint8_t mtype;
	uint16_t length;
	uint32_t mac;
	uint16_t crc;
} __attribute__ ((packed));

struct message_t {
	struct mheader_t header;
	char * payload;
	uint16_t payload_len;
} __attribute__ ((packed));

void push_char(int fd, char *buffer, char in);
int main(int argc, char *argv[])
{
	int fd = -1;
	int err = 0;
	uint32_t baud;
	char data[500];
	char inchar;
	int i = 0;
	
	struct mheader_t* msg_header = (struct mheader_t*)data;
	
	speed_t devicespeed = B57600;
	struct termios serialmode;
	struct termios confirmation;

	if (argc < 3)
	{
		printf("client <device> <speed>\n");
		return -1;
	}
	if (sscanf(argv[2], "%u", &baud) != 1)
	{
		printf("Incorrect baud setting please try again\n");
		return -1;
	}
	switch(baud)
	{
		case 2400: devicespeed = B2400; break;
		case 4800: devicespeed = B4800; break;
		case 9600: devicespeed = B9600; break;
		case 19200: devicespeed = B19200; break;
		case 38400: devicespeed = B38400; break;
		case 57600: devicespeed = B57600; break;
		case 115200: devicespeed = B115200; break;
		case 230400: devicespeed = B230400; break;
		case 500000: devicespeed = B500000; break;
		default: printf("Incorrect baud setting please try again\n");
			return -1;
	}
	// open
	fd = open(argv[1], O_RDWR);
	if (fd < 0)
	{
		printf("open error\n");
		return -1;
	}
	// flush
	err = tcflush(fd, TCIOFLUSH);
	if (err < 0)
	{
		printf("flush error\n");
		close(fd);
		return -1;
	}
	// get terminal modes
	err = tcgetattr(fd, &serialmode);
	if (err < 0)
	{
		printf("tcgetattr error\n");
		close(fd);
		return -1;
	}
	// configure to raw
	serialmode.c_iflag &= ~(IGNBRK | BRKINT | IGNPAR | PARMRK | INPCK | ISTRIP | INLCR | IGNCR | ICRNL | IXON | IXOFF | IUCLC | IXANY | IMAXBEL | IUTF8);
	serialmode.c_oflag &= ~(OPOST | OLCUC | OCRNL | ONLCR | ONOCR | ONLRET | OFILL | OFDEL );
	serialmode.c_lflag &= ~(ISIG | ICANON | IEXTEN | ECHO | ECHOE | ECHOK | ECHONL | NOFLSH);
	serialmode.c_cflag &= ~(CSIZE | PARENB | PARODD | HUPCL | CSTOPB);
	serialmode.c_cflag |= (CS8 | CREAD | CLOCAL);
	err = cfsetispeed(&serialmode, devicespeed);
	if (err < 0)
	{
		printf("cfsetispeed error\n");
		close(fd);
		return -1;
	}
	err = cfsetospeed(&serialmode, devicespeed);
	if (err < 0)
	{
		printf("cfsetospeed error\n");
		close(fd);
		return -1;
	}
	// set new terminal modes
	err = tcsetattr(fd, TCSANOW, &serialmode);
	if (err < 0)
	{
		printf("tcsetattr error \n");
		close(fd);
		return -1;
	}
	for (;;)
	{
		if (read(fd, &inchar, 1) < 0)
		{
			printf("There was a read error.\n");
			exit(0);
		}
		push_char(fd, data, inchar);
	}
}

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
	printf("mac: %d\n", msg_header->mac);
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
void send_packet(int fd, struct message_t * msg) {
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
	write(fd, data, sizeof(struct mheader_t) + 2 + msg->payload_len + 2);
}
	
void process_packet(int fd, char *buffer, int length) {
	struct mheader_t *header;
	struct message_t msg;
	unsigned long tmp;
	static uint8_t addr = 0;
	static uint32_t mac = 128;
	
	print_packet(buffer, length);
	header = (struct mheader_t *)(buffer + 1);
	
	if (addr == 0) {
		if (header->destination == 0) {
			if (header->mtype == 72 && header->mac == 0) {
				// some random delay
				msg.header.destination = 1;
				msg.header.mac = mac;
				msg.header.mtype = 73;
				msg.payload = "NEED IP";
				msg.payload_len = 7;
				send_packet(fd, &msg);
			}
			if (header->mtype == 74 && header->mac == mac) {
				msg.header.destination = 1;
				msg.header.mac = mac;
				msg.header.mtype = 75;
				msg.payload = "ACK";
				msg.payload_len = 3;
				send_packet(fd, &msg);
				tmp = strtoul(buffer + 2 + sizeof(struct mheader_t), NULL, 10);
				addr = 0xFF&tmp;
				printf("THE ADDR: %d\n", addr);
			}
		}
	}
	else {
		printf("Got addr!\n");
		if (header->destination == addr && header->mac == mac) {
			msg.header.destination = 1;
			msg.header.mac = 128;
			msg.header.mtype = 2;
			msg.payload = "HAHAHA";
			msg.payload_len = 6;
			send_packet(fd, &msg);
		}
		else if (header->destination == 0 && header->mtype == 58 && header->mac == 0) {
			// received a reset address request. 
			addr = 0;
		}
	}
	
}

void push_char(int fd, char *buffer, char in) {
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
			
		printf("header checksum: %#x\n", checksum);
		if (checksum != msg_header->crc) {
			goto remove_first;
		}
	}
	if (end > sizeof(struct mheader_t) + 2) {
		if (end - (sizeof(struct mheader_t) + 2) == msg_header->length) {
			checksum = 0xffff;
			for (int i = 0; i < end - 2; i++)
				checksum = crc16_update(checksum, buffer[i]);
			printf("comp checksum: %#x\n", checksum);
			uint16_t * crc = (uint16_t*)(buffer + (end - 2));
			printf("direct checksum: %#x\n", *crc);
			if (checksum == *crc)
			{
				process_packet(fd, buffer, end);
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
