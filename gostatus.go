package main

import (
    "encoding/json"
    "fmt"
    "time"
    "github.com/shirou/gopsutil/load"
    "github.com/shirou/gopsutil/net"
    "github.com/shirou/gopsutil/disk"
    "strconv"
    "os/exec"
    "math"
    "net/http"
)

// SIGSTOP = 19, SIGCONT = 18

type Header struct {
    Version int `json:"version"`;
    StopSignal int `json:"stop_signal,omitempty"`;
    ContSignal int `json:"cont_signal,omitempty"`;
    ClickEvents bool `json:"click_events,omitempty"`;
}

type Block struct {
    FullText string `json:"full_text"`;
    ShortText string `json:"short_text,omitempty"`;
    Color string `json:"color,omitempty"`;
    MinWidth string `json:"min_width,omitempty"`;
    Align int `json:"align,omitempty"`;
    Name string `json:"name"`;
    Instance string `json:"instance,omitempty"`;
    Urgent bool `json:"urgent,omitempty"`;
    Separator bool `json:"separator,omitempty"`;
    SeparatorBlockWidth int `json:"separator_block_width,omitempty"`;
    Markup string `json:"markup,omitempty"`;
}

const OKAY = "#00AA00";
const WARN = "#FFA500";
const BAD = "#FF0000";
const GRAY = "#CCCCCC";

func PrintLine(line []Block) {
    json_bytes, _ := json.Marshal(line);
    fmt.Println(string(json_bytes)+",");
}

func FormatGigabytes(n uint64) string {
    return strconv.FormatFloat(float64(n)/1024/1024/1024, 'f', 2, 64)
}

var (
    anim_frame = 0;
    anim_text = "";
    mode = "standard";
)

type Notification struct {
    Sender string `json:"sender"`;
    Text string `json:"text"`;
}

func notifyd () {
    http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
        decoder := json.NewDecoder(r.Body)
        var notification Notification;
        decoder.Decode(&notification);
        anim_text = fmt.Sprintf("[%s] %s", notification.Sender, notification.Text);
        anim_frame = 0;
        mode = "animation";
    })

    http.ListenAndServe("127.0.0.1:7612", nil)
}

func main () {
    header, _ := json.Marshal(Header{Version: 1})
    fmt.Println(string(header));
    fmt.Println("[");
    go notifyd();
    for {
        if mode == "standard" {
            line := make([]Block, 0);

            // TIME block
            t := time.Now();
            const t_format = "2006-01-02 15:04:05";
            time_block := Block{
                Name: "time",
                FullText: t.Format(t_format),
                Color: GRAY,
            }

            // LOAD block
            load, _ := load.LoadAvg()
            load_block := Block{
                Name: "load",
                FullText: strconv.FormatFloat(load.Load1, 'f', 2, 64),
            }
            if load.Load1 > 2 {
                load_block.Color = BAD;
            } else if load.Load1 > 1 {
                load_block.Color = WARN;
            } else {
                load_block.Color = OKAY;
            }

            // NET block
            net_blocks := make([]Block, 0);
            interfaces, _ := net.NetInterfaces();
            for _, iface := range interfaces {
                if iface.Name == "lo" {
                    continue
                }
                text := iface.Name + ": " + iface.Addrs[0].Addr;
                net_blocks = append(net_blocks, Block{
                    Name: "iface",
                    FullText: text,
                    Instance: iface.Name,
                    Color: GRAY,
                });
            }

            // HDD block
            root_stat, _ := disk.DiskUsage("/");
            home_stat, _ := disk.DiskUsage("/home");
            data_stat, _ := disk.DiskUsage("/media/me/data");

            root_block := Block{
                Name: "hdd",
                Instance: "root",
                FullText: "/: " + FormatGigabytes(root_stat.Free) + "GB",
            }
            if root_stat.Free > 5*1024*1024*1024 {
                root_block.Color = OKAY;
            } else if root_stat.Free > 2.5*1024*1024*1024 {
                root_block.Color = WARN;
            } else {
                root_block.Color = BAD;
            }

            home_block := Block{
                Name: "hdd",
                Instance: "root",
                FullText: "/home: " + FormatGigabytes(home_stat.Free) + "GB",
            }
            if home_stat.Free > 20*1024*1024*1024 {
                home_block.Color = OKAY;
            } else if home_stat.Free > 10*1024*1024*1024 {
                home_block.Color = WARN;
            } else {
                home_block.Color = BAD;
            }

            data_block := Block{
                Name: "hdd",
                Instance: "data",
                FullText: "data: " + FormatGigabytes(data_stat.Free) + "GB",
            }
            if data_stat.Free > 30*1024*1024*1024 {
                data_block.Color = OKAY;
            } else if data_stat.Free > 15*1024*1024*1024 {
                data_block.Color = WARN;
            } else {
                data_block.Color = BAD;
            }

            // Headphones block
            headphones := Block {
                Name: "headphones",
                FullText: "Headphones ",
            }
            if exec.Command("sh", "-c", "pacmd list-sinks | grep DR-BTN200").Run() == nil {
                headphones.FullText += "connected";
                headphones.Color = OKAY;
            } else {
                headphones.FullText += "disconnected";
                headphones.Color = BAD;
            }


            line = append(line, headphones);
            line = append(line, root_block);
            line = append(line, home_block);
            line = append(line, data_block);
            for _, block := range net_blocks {
                line = append(line, block);
            }
            line = append(line, load_block);
            line = append(line, time_block);

            PrintLine(line);
            time.Sleep(time.Second * 1);
        } else if mode == "animation" {
            text := anim_text[0:anim_frame];
            anim_frame++;
            PrintLine([]Block{Block{Name:"anim",FullText:text,Instance: strconv.Itoa(anim_frame)}});
            if anim_frame > len(anim_text) {
                mode = "standard";
                anim_frame = 0;
                time.Sleep(time.Second * 3);
            } else {
                time.Sleep(time.Millisecond * time.Duration(math.Max(35, float64(40 - len(anim_text)))));
            }
        }
    }
}

