import customtkinter as ctk
from tkinter import filedialog, messagebox, simpledialog
import os


class TDLGUI(ctk.CTk):
    def __init__(self):
        super().__init__()
        self.title("TDL 图形界面工具")
        self.geometry("960x1000")
        ctk.set_appearance_mode("dark")
        ctk.set_default_color_theme("blue")

        self.download_dir = ctk.StringVar()
        self.merge_mode = ctk.BooleanVar()
        self.extra_args = ctk.StringVar()
        self.command_output = ctk.StringVar()

        # 中文解释 + 参数
        self.check_options = {
            "反序下载 (--desc)": "--desc",
            "自动重命名扩展名 (--rewrite-ext)": "--rewrite-ext",
            "组合消息识别 (--group)": "--group",
            "跳过相同文件 (--skip-same)": "--skip-same",
            "恢复下载 (--continue)": "--continue",
            "重新开始下载 (--restart)": "--restart",
            "HTTP 文件服务器 (--serve)": "--serve",
        }

        self.tooltips = {
            "反序下载 (--desc)": "从最新消息开始下载，影响恢复下载行为",
            "自动重命名扩展名 (--rewrite-ext)": "按 MIME 类型重命名扩展名（如 .apk -> .zip）",
            "组合消息识别 (--group)": "下载相册或组合消息中的所有文件",
            "跳过相同文件 (--skip-same)": "跳过已存在且大小一致的文件",
            "恢复下载 (--continue)": "无交互方式恢复上次下载任务",
            "重新开始下载 (--restart)": "无交互方式重新开始任务",
            "HTTP 文件服务器 (--serve)": "本地开启 HTTP 服务用于外部工具下载",
        }

        self.check_vars = {}

        self.create_widgets()

    def create_widgets(self):
        # 顶部安装按钮
        install_frame = ctk.CTkFrame(self)
        install_frame.pack(fill="x", padx=10, pady=5)
        ctk.CTkButton(
            install_frame, text="安装 TDL 命令行工具", command=self.install_cli
        ).pack(anchor="w", padx=5, pady=5)

        login_frame = ctk.CTkFrame(self)
        login_frame.pack(fill="x", padx=10, pady=5)

        ctk.CTkLabel(login_frame, text="登录方式：").pack(side="left", padx=5)
        ctk.CTkButton(
            login_frame,
            text="默认路径登录",
            command=lambda: self.run_in_terminal("tdl login"),
        ).pack(side="left", padx=5)
        ctk.CTkButton(
            login_frame, text="密码登录", command=self.login_with_passcode
        ).pack(side="left", padx=5)
        ctk.CTkButton(
            login_frame,
            text="二维码登录",
            command=lambda: self.run_in_terminal("tdl login -T qr"),
        ).pack(side="left", padx=5)
        ctk.CTkButton(
            login_frame,
            text="验证码登录",
            command=lambda: self.run_in_terminal("tdl login -T code"),
        ).pack(side="left", padx=5)

        download_frame = ctk.CTkFrame(self)
        download_frame.pack(fill="both", expand=True, padx=10, pady=5)

        ctk.CTkLabel(
            download_frame, text="批量下载 URL（每行一个）默认包含 --takeout 参数"
        ).pack(anchor="w", padx=5)
        self.url_text = ctk.CTkTextbox(download_frame, height=100)
        self.url_text.pack(fill="both", expand=True, padx=5, pady=5)

        btn_frame = ctk.CTkFrame(download_frame)
        btn_frame.pack(anchor="w", padx=5, pady=5)
        ctk.CTkButton(btn_frame, text="下载 URL", command=self.download_url).pack(
            side="left", padx=5
        )
        ctk.CTkButton(
            btn_frame, text="下载 JSON", command=self.select_and_download_json
        ).pack(side="left", padx=5)

        # 参数复选 + 文本框
        option_frame = ctk.CTkFrame(self)
        option_frame.pack(fill="x", padx=10, pady=10)

        ctk.CTkLabel(option_frame, text="常用参数（鼠标悬停可查看说明）：").pack(
            anchor="w", padx=5
        )
        for label, flag in self.check_options.items():
            var = ctk.BooleanVar()
            cb = ctk.CTkCheckBox(option_frame, text=label, variable=var)
            cb.pack(anchor="w", padx=15, pady=2)
            cb.bind(
                "<Enter>", lambda e, text=self.tooltips[label]: self.show_tooltip(text)
            )
            cb.bind("<Leave>", lambda e: self.hide_tooltip())
            self.check_vars[flag] = var

        self.tooltip_label = ctk.CTkLabel(option_frame, text="", text_color="gray")
        self.tooltip_label.pack(anchor="w", padx=20, pady=5)

        ctk.CTkLabel(option_frame, text="附加参数：").pack(
            anchor="w", padx=5, pady=(10, 0)
        )
        ctk.CTkEntry(option_frame, textvariable=self.extra_args, width=600).pack(
            anchor="w", padx=15, pady=5
        )

        ctk.CTkCheckBox(
            option_frame,
            text="合并模式（一个窗口下载全部链接）",
            variable=self.merge_mode,
        ).pack(anchor="w", padx=15, pady=5)
        ctk.CTkButton(
            option_frame, text="选择下载目录", command=self.select_download_dir
        ).pack(anchor="w", padx=15, pady=5)
        ctk.CTkLabel(
            option_frame, textvariable=self.download_dir, text_color="lightblue"
        ).pack(anchor="w", padx=15)

        output_frame = ctk.CTkFrame(self)
        output_frame.pack(fill="both", expand=False, padx=10, pady=10)
        ctk.CTkLabel(output_frame, text="生成命令（将执行）：").pack(anchor="w", padx=5)
        self.output_text = ctk.CTkTextbox(output_frame, height=120)
        self.output_text.pack(fill="both", expand=True, padx=5, pady=5)
        ctk.CTkButton(output_frame, text="复制命令", command=self.copy_cmd).pack(
            anchor="e", padx=10, pady=5
        )

    def install_cli(self):
        cmd = "iwr -useb https://docs.iyear.me/tdl/install.ps1 | iex"
        self.clipboard_clear()
        self.clipboard_append(cmd)
        messagebox.showinfo(
            "手动安装提示",
            "请以管理员身份打开 PowerShell，粘贴并运行以下命令：\n\n"
            + cmd
            + "\n\n已自动复制到剪贴板。",
        )

    def show_tooltip(self, text):
        self.tooltip_label.configure(text=f"说明：{text}")

    def hide_tooltip(self):
        self.tooltip_label.configure(text="")

    def run_in_terminal(self, cmd):
        os.system(f'start cmd /k "{cmd}"')

    def login_with_passcode(self):
        pw = simpledialog.askstring("输入密码", "请输入本地密码：")
        if pw:
            self.run_in_terminal(f"tdl login -p {pw}")

    def select_download_dir(self):
        path = filedialog.askdirectory()
        if path:
            self.download_dir.set(path)

    def copy_cmd(self):
        cmd = self.output_text.get("1.0", "end").strip()
        self.clipboard_clear()
        self.clipboard_append(cmd)
        messagebox.showinfo("提示", "命令已复制到剪贴板")

    def build_common_args(self):
        args = ["--takeout"]
        for flag, var in self.check_vars.items():
            if var.get():
                args.append(flag)
        extra = self.extra_args.get()
        if extra:
            args.append(extra)
        if self.download_dir.get():
            args.append(f'-d "{self.download_dir.get()}"')
        return " ".join(args)

    def download_url(self):
        urls = self.url_text.get("1.0", "end").strip().splitlines()
        if not urls:
            messagebox.showwarning("提示", "请输入至少一个下载链接")
            return

        common_args = self.build_common_args()

        if self.merge_mode.get():
            cmd = "tdl dl"
            for url in urls:
                if url.strip():
                    cmd += f" -u {url.strip()}"
            cmd += f" {common_args}"
            self.output_text.delete("1.0", "end")
            self.output_text.insert("1.0", cmd)
            self.run_in_terminal(cmd)
        else:
            cmds = [
                f"tdl dl -u {url.strip()} {common_args}" for url in urls if url.strip()
            ]
            cmd_str = "\n".join(cmds)
            self.output_text.delete("1.0", "end")
            self.output_text.insert("1.0", cmd_str)
            for cmd in cmds:
                self.run_in_terminal(cmd)

    def select_and_download_json(self):
        files = filedialog.askopenfilenames(filetypes=[("JSON 文件", "*.json")])
        if not files:
            return
        cmd = "tdl dl"
        for f in files:
            cmd += f' -f "{f}"'
        cmd += f" {self.build_common_args()}"
        self.output_text.delete("1.0", "end")
        self.output_text.insert("1.0", cmd)
        self.run_in_terminal(cmd)


if __name__ == "__main__":
    TDLGUI().mainloop()
