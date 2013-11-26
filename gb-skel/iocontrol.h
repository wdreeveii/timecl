
// io control code
#ifndef _IOCONTROL_H
#define _IOCONTROL_H

#define NUM_PORTS 14
#define PORT_OUTPUT 0
#define PORT_BINPUT 2
#define PORT_AINPUT 4

void io_init();

uint8_t io_num_ports();
uint8_t io_get_type(uint8_t port);

uint8_t io_read(uint8_t index);
uint16_t io_aread(uint8_t index);

void iocontrol(unsigned char port, unsigned char state);
void ioflip(unsigned char port);

#endif
