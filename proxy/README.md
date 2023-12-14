# connector设备服务

## 1、协议类型
### 1.1、物模型协议

|类型|topic|消息格式|描述|
|--|--|--|--|
|设备属性上报|sys/{deviceSN}/thing/json/property/post|{"id": "123","version": "1.0","params": {"Power": {"value": "on","time": 1524448722000},"WF": {"value": 23.6,"time": 1524448722000}}}|设备产生的属性，可以支持批量属性上报
|设备属性上报回复|sys/{deviceSN}/thing/json/property/post_reply|{"code": 0,"msg": "","id": "123","version": "1.0","data": {}}|云端回复设备属性上报成功
|设备事件上报|sys/{deviceSN}/thing/json/event/${tsl.event.identifier}/post|{"id": "123","version": "1.0","params": {"value": {"Power": "on","WF": "2"},"time": 1524448722000}}|设备产生的事件，支持单个事件上报
|设备事件上报回复|sys/{deviceSN}/thing/json/event/${tsl.event.identifier}/post_reply|{"code": 0,"msg": "","id": "123","version": "1.0","data": {}}|云端回复设备事件上报成功
|云端服务调用|sys/{deviceSN}/thing/json/service/${tsl.service.identifier}/call|{"id": "123","version": "1.0","params": {"Power": "on","WF": "2"}}|云端下发服务调用指令
|云端服务调用返回|sys/{deviceSN}/thing/json/service/${tsl.service.identifier}/call_reply|{"code": 0,"msg": "","id": "123","version": "1.0","data": {}}|设备回复云端服务调用结果
|云端属性设置|sys/{deviceSN}/thing/json/service/property/set|{"id": "123","version": "1.0","params": {"Power": "on","WF": "2"}}|云端下发属性设置指令
|云端属性设置返回|sys/{deviceSN}/thing/json/service/property/set_reply|{"code": 0,"msg": "","id": "123","version": "1.0","data": {}}|设备回复设备属性设置结果

### 1.2、用户透传协议
|类型|topic|消息格式|描述|
|--|--|--|--|
|设备属性上报|sys/{deviceSN}/thing/raw/property|rawdata|设备产生的属性，可以支持批量属性上报
|设备属性上报回复|sys/{deviceSN}/thing/raw/property_reply|rawdata|云端回复设备属性上报成功
|设备事件上报|sys/{deviceSN}/thing/raw/event/${tsl.event.identifier}/post|rawdata|设备产生的事件，支持单个事件上报
|设备事件上报回复|sys/{deviceSN}/thing/raw/event/${tsl.event.identifier}/post_reply|rawdata|云端回复设备事件上报成功
|云端服务调用|sys/{deviceSN}/thing/raw/service/${tsl.service.identifier}/call|rawdata|云端下发服务调用指令
|云端服务调用返回|sys/{deviceSN}/thing/raw/service/${tsl.service.identifier}/call_reply|rawdata|设备回复云端服务调用结果
|云端属性设置|sys/{deviceSN}/thing/raw/service/property/set|rawdata|云端下发属性设置指令
|云端属性设置返回|sys/{deviceSN}/thing/raw/service/property/set_reply|rawdata|设备回复设备属性设置结果
