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

另外，我用python语言也实现了一版，结果发现非并发的python运行得相当快。可能与python中使用了yield有关。

我在一个10w行的Repo中测试性能，得到以下结果：
* golang非并发：49秒
* golang并发：7秒
* python非并发：35秒
* python并发
