
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
	printf("crc: %d\n", msg_header->crc);
	printf("crc hex: %#x\n", msg_header->crc);
	for (int i = sizeof(struct mheader_t) +2; i < length; i++)
		printf("%c", *(buffer + i));
	printf("\n");
	for (int i = sizeof(struct mheader_t) +2; i < length; i++)
		printf("%#x ", *(buffer + i));
	printf("\n");
	printf("payload crc: %#x\n", *((uint16_t*)(buffer + (length - 2))) );
}

uint16_t header_crc(char * buffer)
{
	uint16_t checksum = 0xffff;
	for (int i = 0; i < sizeof(struct mheader_t) - 1; i++)
		checksum = crc16_update(checksum, buffer[i]);
	return checksum;
}

void process_packet(int fd, char *buffer, int length) {
	char data[sizeof(struct mheader_t) + 2 + 6 + 2];
	struct mheader_t * header;
	uint16_t checksum = 0xffff;
	print_packet(buffer, length);
	
	data[0] = 'A';
	data[sizeof(struct mheader_t) + 1] = 'A';
	
	memcpy((data + sizeof(struct mheader_t) + 2), "HAHAHA", 6);
	
	header = ((struct mheader_t *)(data + 1));
	header->destination = 2;
	header->mac = 128;
	header->mtype = 2;
	header->length = 6 + 2;
	header->crc = header_crc(data);
	
	for (int i = 0; i < sizeof(struct mheader_t) + 2 + 6; i++)
		checksum = crc16_update(checksum, data[i]);
	
	*((uint16_t *)(data + sizeof(struct mheader_t) + 2 + 6)) = checksum;
	write(fd, data, sizeof(struct mheader_t) + 2 + 8);
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
