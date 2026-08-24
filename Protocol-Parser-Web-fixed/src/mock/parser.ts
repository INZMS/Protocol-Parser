export const mockParseResult = {
  protocol: "JT-808",
  messageId: "0200",
  messageName: "位置上传",
  length: 69,

  fields: [
    { name: "消息ID", offset: 0, len: 2, hex: "0200", value: "位置上传", desc: "消息类型" },
    { name: "消息体属性", offset: 2, len: 2, hex: "0045", value: "定位数据", desc: "属性字段" },
    { name: "终端手机号", offset: 4, len: 6, hex: "013800138000", value: "13800138000", desc: "终端号码" },
    { name: "流水号", offset: 10, len: 2, hex: "0001", value: "1", desc: "消息流水号" },
    { name: "报警标志", offset: 14, len: 4, hex: "C8001B5E", value: "正常", desc: "报警状态" },
    { name: "状态", offset: 18, len: 4, hex: "E3003C18", value: "ACC开,北斗定位", desc: "状态位" },
    { name: "纬度", offset: 22, len: 4, hex: "18071D0C", value: "31.230400°", desc: "纬度(度)" },
    { name: "经度", offset: 26, len: 4, hex: "0A1E0000", value: "121.473700°", desc: "经度(度)" },
    { name: "速度", offset: 30, len: 2, hex: "003C", value: "60 km/h", desc: "单位:0.1km/h" },
    { name: "方向", offset: 32, len: 2, hex: "005A", value: "90°", desc: "单位:0.1°" },
    { name: "定位时间", offset: 34, len: 6, hex: "260729103025", value: "2026-07-29 10:30:25", desc: "YYMMDDhhmmss" }
  ],

  rawHex: "7E020000450100000000013800138000018071D0C0A1E0000003C005A2607291030257E",

  json: {
    protocol: "JT-808",
    messageId: "0200",
    messageName: "位置上传",
    data: {
      terminalNo: "13800138000",
      latitude: 31.2304,
      longitude: 121.4737,
      speed: 60,
      direction: 90,
      time: "2026-07-29 10:30:25"
    }
  }
};

export const mockExamples = [
  { label: "2929位置上报(80)", protocol: "2929", hex: "292980013B14941494210417185424022409531141642500000000FF000000FFF750FFFFFFFFFFFFFFFF00001200243436303B30303B393336393B343934320005000429680000040008045F000F00A359475F434430315F52322E373B000600A50000001000060089FFFFFFFF002400A901CC00032499134E1C249913121C249911EF1C000000000000000000000000000000007000B90532633A36313A30343A37393A64333A39632C2D36372C63633A30383A66623A39383A62373A39352C2D37322C30383A39623A34623A39643A39623A35312C2D37332C39633A32313A36613A65343A31353A65382C2D37382C39633A61363A31353A39663A64303A62302C2D3833000600C500001010001600FB3839383630343132313031383730383434363635000500AE0300050005F0000100050009F00121041718590700D50D" },
  { label: "JT-808位置上报(0200)", protocol: "JT-808", hex: "7E020000220138001380000001000000000000000301DAE80007383280000C0258005A24082412345630011F310108547E" },
  { label: "VDF定时定位上报(BA)", protocol: "VDF", hex: "2A48513230313836313830333137383939393030322C42412641313131313138323233333339363831313335363532313136303031363039303731382642303130303030303030302646303030302652313830372657303030303030323926493534363030303234393530453131353132343935304531323534323439353133423334323234393531323233333532363233313044373335264B34303130302654393523" },
  { label: "中心确认(21)", protocol: "2929", hex: "2929210005D084C4B40D" },
  { label: "申请设置参数(D8)", protocol: "2929", hex: "2929D8000601020304DA0D" }
];

export const mockHistory = [
  { id: 1, time: "2026-07-26 10:20:30", protocol: "JT-808", messageId: "0200", messageName: "位置上传", length: 69 },
  { id: 2, time: "2026-07-26 10:15:22", protocol: "JT-808", messageId: "0002", messageName: "终端心跳", length: 13 },
  { id: 3, time: "2026-07-26 10:10:05", protocol: "JT-808", messageId: "0100", messageName: "终端注册", length: 32 },
  { id: 4, time: "2026-07-25 15:43:11", protocol: "2929协议", messageId: "8100", messageName: "终端登录应答", length: 25 },
  { id: 5, time: "2026-07-25 15:40:02", protocol: "2929协议", messageId: "8102", messageName: "位置上传报文", length: 56 },
  { id: 6, time: "2026-07-25 09:12:47", protocol: "JT-808", messageId: "0704", messageName: "位置批量上传", length: 128 },
  { id: 7, time: "2026-07-24 21:03:19", protocol: "JT-808", messageId: "0200", messageName: "位置上传", length: 69 },
  { id: 8, time: "2026-07-24 18:55:40", protocol: "JT-808", messageId: "0002", messageName: "终端心跳", length: 13 },
  { id: 9, time: "2026-07-24 11:30:08", protocol: "2929协议", messageId: "8100", messageName: "终端登录应答", length: 25 },
  { id: 10, time: "2026-07-23 20:47:53", protocol: "JT-808", messageId: "0801", messageName: "报警信息", length: 41 },
  { id: 11, time: "2026-07-23 14:22:16", protocol: "JT-808", messageId: "0100", messageName: "终端注册", length: 32 },
  { id: 12, time: "2026-07-23 09:05:33", protocol: "2929协议", messageId: "8102", messageName: "位置上传报文", length: 56 }
];

export const mockRecordDetails: Record<number, any> = {


    // 1 位置上传
    1: {
        ...mockParseResult
    },



    // 2 终端心跳
    2: {

        protocol: "JT-808",

        messageId: "0002",

        messageName: "终端心跳",

        length: 13,


        rawHex:
            "7E00020000000000010000007E",


        fields: [

            {
                name:"消息ID",
                offset:0,
                len:2,
                hex:"0002",
                value:"终端心跳",
                desc:"消息类型"
            },


            {
                name:"消息体属性",
                offset:2,
                len:2,
                hex:"0000",
                value:"无消息体",
                desc:"属性字段"
            },


            {
                name:"流水号",
                offset:4,
                len:2,
                hex:"0001",
                value:"1",
                desc:"消息流水"
            }


        ],


        json:{


            protocol:"JT-808",

            messageId:"0002",

            messageName:"终端心跳",


            data:{


                heartbeat:true,


                time:"2026-07-26 10:15:22"


            }


        }

    },





    // 3 终端注册
    3: {


        protocol:"JT-808",

        messageId:"0100",

        messageName:"终端注册",

        length:32,


        rawHex:
        "7E010000000000320034303030303030303030307E",


        fields:[


            {
                name:"消息ID",
                offset:0,
                len:2,
                hex:"0100",
                value:"终端注册",
                desc:"消息类型"
            },


            {
                name:"终端手机号",
                offset:4,
                len:6,
                hex:"340303030303",
                value:"34000000000",
                desc:"终端号码"
            },


            {
                name:"制造商",
                offset:10,
                len:5,
                hex:"ABCDEF",
                value:"测试厂家",
                desc:"厂商编号"
            }


        ],



        json:{


            protocol:"JT-808",

            messageId:"0100",

            messageName:"终端注册",


            data:{


                terminalNo:"34000000000",

                manufacturer:"测试厂家"


            }


        }

    },







    // 4 2929登录应答
    4:{


        protocol:"2929协议",

        messageId:"8100",

        messageName:"终端登录应答",

        length:25,


        rawHex:
        "292981000000000000000000000000000000",


        fields:[


            {
                name:"消息ID",
                offset:0,
                len:2,
                hex:"8100",
                value:"登录应答",
                desc:"消息类型"
            },


            {
                name:"结果",
                offset:2,
                len:1,
                hex:"00",
                value:"成功",
                desc:"应答结果"
            }


        ],



        json:{


            protocol:"2929协议",

            messageId:"8100",

            messageName:"终端登录应答",


            data:{


                result:"success"


            }


        }


    },








    // 5 2929位置上传
    5:{


        protocol:"2929协议",

        messageId:"8102",

        messageName:"位置上传报文",

        length:56,


        rawHex:
        "292981020000000000000000000000",


        fields:[


            {
                name:"消息ID",
                offset:0,
                len:2,
                hex:"8102",
                value:"位置上传",
                desc:"消息类型"
            },


            {
                name:"纬度",
                offset:10,
                len:4,
                hex:"18071D0C",
                value:"31.230400",
                desc:"纬度"
            },


            {
                name:"经度",
                offset:14,
                len:4,
                hex:"0A1E0000",
                value:"121.473700",
                desc:"经度"
            }


        ],



        json:{


            protocol:"2929协议",

            messageId:"8102",

            messageName:"位置上传报文",


            data:{


                latitude:31.2304,

                longitude:121.4737


            }


        }


    }

};
