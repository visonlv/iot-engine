[TOC]



# 1、物模型基本结构
```
{
    "properties": {
        "speed": {
            "name": "行驶速度",
            "description": "描述",
            "unit": "km/h",
            "type": "float",
            "min": 0.0,
            "max": 10.0,
            "writable": true
        },
        "light": {
            "name": "灯光开关",
            "description": "描述",
            "type": "bool",
            "writable": true
        },
        "fan_level": {
            "name": "风扇级别",
            "description": "描述",
            "type": "int",
            "options": [
                1,
                2,
                3,
                4,
                5
            ],
            "writable": true
        }
    },
    "services": {
        "selfClean": {
            "name": "自清洗",
            "description": "",
            "parameters": [
                {
                    "name": "level",
                    "description": "清洗力度",
                    "type": "int",
                    "options": [
                        1,
                        2,
                        3
                    ]
                }
            ]
        }
    },
    "events": {
        "selfCleanFinished": {
            "name": "自清洗完成",
            "description": "",
            "parameters": [
                {
                    "name": "cost",
                    "description": "清洗耗时",
                    "type": "int"
                }
            ]
        }
    }
}
```