use std::error::Error;
use std::fmt;

use crossbeam_channel::{bounded, Receiver, select};
use ctrlc;
use rppal::gpio::{Gpio, Trigger};
use rppal::i2c::I2c;
use rppal::system::DeviceInfo;

#[derive(Debug)]
struct MyError {
    details: String
}

impl MyError {
    fn new(msg: &str) -> MyError {
        MyError{details: msg.to_string()}
    }
}

impl fmt::Display for MyError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f,"{}",self.details)
    }
}

impl Error for MyError {
    fn description(&self) -> &str {
        &self.details
    }
}

fn ctrl_channel() -> Result<Receiver<()>, ctrlc::Error> {
	let (sender, receiver) = bounded(100);
	ctrlc::set_handler(move || {
		let _ = sender.send(());
	})?;

	Ok(receiver)
}

const GPIO_GEIGER: u8 = 4;
const I2C_ADDRESS: u16 = 0x19;
const I2C_START: [u8; 1] = [0x71];
const I2C_STOP: [u8; 1] = [0x00];

fn main() -> Result<(), Box<dyn Error>> {
	let ctrl_c_events = ctrl_channel()?;

	println!("Starting up geiger on a {}.", DeviceInfo::new()?.model());
	let mut i2c = I2c::new()?;
	i2c.set_slave_address(I2C_ADDRESS)?;
	println!("sending geiger start");
	let written = i2c.write(&I2C_START)?;
	if written != 1 {
		return Err(Box::new(MyError::new("start not written")));
	}

	let gpio = Gpio::new()?;

	let mut counter = 0;
	let particle = move |level| {
		counter = counter + 1;
		println!("particle {} {}", level, counter);
	};
	let mut pin = gpio.get(GPIO_GEIGER)?.into_input();
	pin.set_async_interrupt(Trigger::FallingEdge, particle)?;

	loop {
		select! {
			recv(ctrl_c_events) -> _ => {
				println!("sending geiger stop");
				let written = i2c.write(&I2C_STOP)?;
				if written != 1 {
					return Err(Box::new(MyError::new("stop not written")));
				}
                		break;
            		}
        	}
    	}

	Ok(())
}
