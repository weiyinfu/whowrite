# 快速查看代码是谁写的

安装方式
```sh
git clone git@github.com:weiyinfu/whowrite.git
cd whowrite-master
go install
```

在有些垃圾公司，考察员工绩效的时候会统计每个员工贡献的代码行数。  
统计一个Repo每行代码是谁写的，只需要使用git blame命令就可以查看，例如`git blame haha.py`。这个问题其实就转化为迭代一个目录下各个文件代码是谁写的，多次调用git blame。  

这个问题是一道不错的练习题，涉及到的库包括：
* 文件IO，读取目录、读取文件内容
* 开启进程执行命令，并拿到输出内容
* 正则表达式从文本中提取想要的信息
* golang的并发执行

我在一个10w行的Repo中测试性能，得到以下结果：
* golang非并发：62秒
* python非并发：65秒
* golang并发：25秒
* python并发：16秒