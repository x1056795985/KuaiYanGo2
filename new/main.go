package new

import (
	_ "server/new/app/global"
	. "server/new/app/init"
)

func Main() {
	InitMqttClient()
}
