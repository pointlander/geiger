use std::error::Error;

use crossbeam_channel::{bounded, Receiver, select};
use ctrlc;
use rppal::gpio::{Gpio, Trigger};
use rppal::system::DeviceInfo;

fn ctrl_channel() -> Result<Receiver<()>, ctrlc::Error> {
	let (sender, receiver) = bounded(100);
	ctrlc::set_handler(move || {
		let _ = sender.send(());
	})?;

	Ok(receiver)
}

const GPIO_GEIGER: u8 = 4;

fn main() -> Result<(), Box<dyn Error>> {
	let ctrl_c_events = ctrl_channel()?;

	println!("Starting up geiger on a {}.", DeviceInfo::new()?.model());

	let gpio = Gpio::new()?;

	let mut counter = 0;
	let particle = move |level| {
		counter = counter + 1;
		println!("particle {} {}", level, counter);
	};
	let mut pin = gpio.get(GPIO_GEIGER)?.into_input_pulldown();
	pin.set_async_interrupt(Trigger::FallingEdge, particle)?;

	loop {
		select! {
			recv(ctrl_c_events) -> _ => {
				println!("Exiting...");
                		break;
            		}
        	}
    	}

	Ok(())
}
