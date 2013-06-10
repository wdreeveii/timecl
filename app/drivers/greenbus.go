package drivers

import (
	"os"
	"time"
	"fmt"
	"bytes"
	"encoding/binary"
	//"github.com/robfig/revel"
	//"github.com/mewkiz/pkg/hashutil/crc16"
	//"github.com/robfig/revel/modules/jobs/app/jobs"
	"github.com/wdreeveii/termioslib"
	"timecl/app/network_manager"
)

type GreenBus struct {
    StopChan	chan bool
}

func (d GreenBus) Init(port string) {
    go d.runmaster(port)
}

func (d GreenBus) Stop() {
    d.StopChan <- true
}

func (d GreenBus) Get(cmd network_manager.GetDrvCmd) {
    
}

func (d GreenBus) Set(cmd network_manager.SetDrvCmd) {
    
}

type Header_t struct {
    Destination uint8
    Mtype uint8
    Length uint16
    Mac uint32
    Crc uint16
}

type Message_t struct {
    Header Header_t
    Payload []byte
}

func crc16_update(crc uint16, a byte) uint16 {
    var i int
    crc ^= uint16(a)
    for i = 0; i < 8; i++ {
	if (crc & 1 != 0) {
	    crc = (crc >> 1) ^ 0xA001
	} else {
	    crc = (crc >> 1)
	}
    }
    return crc
}

func ToByte(msg Message_t) []byte {
    b := new(bytes.Buffer)
    b.Write([]byte("A"))
    binary.Write(b, binary.LittleEndian, msg.Header)
    b.Write([]byte("A"))
    b.Write(msg.Payload)
    bytes := b.Bytes()
    var checksum uint16 = 0xffff
    for i := 0; i < len(bytes); i++ {
	checksum = crc16_update(checksum, bytes[i])
    }
    binary.Write(b, binary.LittleEndian, checksum)
    return b.Bytes()
}

func HeaderCRC(msg *Message_t) uint16 {
    msg.Header.Length = uint16(len(msg.Payload) + 2) // add space for the crc
    b := new(bytes.Buffer)
    b.Write([]byte("A"))
    binary.Write(b, binary.LittleEndian, msg.Header)
    bytes := b.Bytes()
    var checksum uint16 = 0xffff
    for i := 0; i < len(bytes) - 2; i++ {
	checksum = crc16_update(checksum, bytes[i])
    }
    return checksum
}

func (d GreenBus) runmaster(port string) {
    var (
	err error
	orig_termios termioslib.Termios
	work_termios termioslib.Termios
	ser *os.File
    )
    
    defer func () {
	fmt.Println("Serial Exiting...")
	if err != nil { 
	    fmt.Println(err)
	}
    }()
    
    ser, err = os.OpenFile(port, os.O_RDWR, 777)
    if err != nil { return }
    
    defer func () {
	fmt.Println("Closing serial port")
	ser.Close()
    }()
    
    
    
    // flush all buffers
    if err = termioslib.Flush(ser.Fd(), termioslib.TCIOFLUSH); err != nil { return }
    
    // save a copy of the original terminal configuration
    if err = termioslib.Getattr(ser.Fd(), &orig_termios); err != nil { return }

    // get a working copy
    if err = termioslib.Getattr(ser.Fd(), &work_termios); err != nil { return }

    work_termios.C_iflag &= ^(termioslib.IGNBRK | termioslib.BRKINT | termioslib.IGNPAR | termioslib.PARMRK | termioslib.INPCK | termioslib.ISTRIP | termioslib.INLCR | termioslib.IGNCR | termioslib.ICRNL | termioslib.IXON | termioslib.IXOFF | termioslib.IXANY | termioslib.IMAXBEL | termioslib.IUTF8)
    work_termios.C_oflag &= ^(termioslib.OPOST | termioslib.ONLCR )
    work_termios.C_lflag &= ^(termioslib.ISIG | termioslib.ICANON | termioslib.IEXTEN | termioslib.ECHO | termioslib.ECHOE | termioslib.ECHOK | termioslib.ECHONL | termioslib.NOFLSH | termioslib.TOSTOP | termioslib.ECHOPRT | termioslib.ECHOCTL | termioslib.ECHOKE)
    work_termios.C_cflag &= ^(termioslib.CSIZE | termioslib.PARENB | termioslib.PARODD | termioslib.HUPCL | termioslib.CSTOPB | termioslib.CRTSCTS)
    work_termios.C_cflag |= (termioslib.CS8 | termioslib.CREAD | termioslib.CLOCAL)
    work_termios.C_cc[termioslib.VMIN] = 1
    work_termios.C_cc[termioslib.VTIME] = 0

    termioslib.Setspeed(&work_termios, termioslib.B57600)
    // set the working copy
    if err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW , &work_termios); err != nil { return }

    // set the settings back to the original when the program exits
    defer func () {
	fmt.Println("Resetting termios")
        err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW, &orig_termios)
    } ()

    r:= make(chan Message_t)
    w:= make(chan Message_t)
    go WriteMessages(ser, w)
    go ReadMessages(ser, r)
    for {
	var msg Message_t
	msg.Payload = []byte("HAHAHA")
	
	msg.Header.Destination = 2
	msg.Header.Mtype = 45
	msg.Header.Mac = 128
	
	w <- msg
	fmt.Println("beat")
	select {
	case <- d.StopChan:
	    fmt.Println("STOPPING DRIVER!!!")
	    return
	case rmsg := <- r:
	    fmt.Printf("read: %#v\n", rmsg)
	    time.Sleep(1*time.Second)
	case <- time.After(1*time.Second):
	    fmt.Printf("read timeout\n")
	    continue
	}
    }
}

func ReadMessages(c *os.File, r chan Message_t) {
    var buffer []byte
    var hsize int = binary.Size(Header_t {})
    in := make([]byte, 1)
    for {
	num_read, err := c.Read(in)
	if err != nil {
	    fmt.Println(err)
	}
	if num_read > 0 {
	    buffer = append(buffer, in[0])
	    var header Header_t
	    if len(buffer) == hsize + 2 {
		if buffer[0] != 'A' || buffer[hsize + 1] != 'A' {
		    buffer = buffer[1:]
		    continue
		}
		var checksum uint16 = 0xffff
		for i:= 0; i < hsize - 1; i++ {
		    checksum = crc16_update(checksum, buffer[i])
		}
		iobuffer := new(bytes.Buffer)
		iobuffer.Write(buffer[1:hsize + 1])
		err = binary.Read(iobuffer, binary.LittleEndian, &header)
		if err != nil {
		    fmt.Println(err)
		}
		
		if checksum != header.Crc {
		    fmt.Println("Header CRC doesn't match")
		    buffer = buffer[1:]
		    continue
		}
	    }
	    if len(buffer) > hsize + 2 {
		iobuffer := new(bytes.Buffer)
		iobuffer.Write(buffer[1:hsize+1])
		err = binary.Read(iobuffer, binary.LittleEndian, &header)
		if err != nil {
		    fmt.Println(err)
		}
		if len(buffer) == hsize + 2 + int(header.Length) {
		    var checksum uint16 = 0xffff
		    for i := 0; i < len(buffer) - 2; i++ {
			checksum = crc16_update(checksum, buffer[i])
		    }
		    if checksum == binary.LittleEndian.Uint16(buffer[len(buffer) - 2:]) {
			var msg Message_t
			msg.Header = header
			msg.Payload = buffer[hsize + 2:len(buffer) - 2]
			r <- msg
		    }
		    buffer = make([]byte, 0)
		}
		if len(buffer) > hsize + 2 + int(header.Length) {
		    buffer = make([]byte, 0)
		}
	    }
	    
	}
    }
}
func WriteMessages(c *os.File, w chan Message_t) {
    var msg Message_t
    for {
	msg = <- w
	msg.Header.Crc = HeaderCRC(&msg)
	fmt.Printf("Msg To: %d Type: %d Length: %d\n", msg.Header.Destination, msg.Header.Mtype, msg.Header.Length)
	_, err := c.Write(ToByte(msg))
	if err != nil {
	    fmt.Println(err)
	    //return
	}
    }
}
    
func init() {
	fmt.Println("blah")
	network_manager.RegisterDriver("greenbus", GreenBus{StopChan: make(chan bool)})
	//go runmaster()
}
