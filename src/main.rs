use std::{
    env,                              // 引入环境变量模块
    net::{SocketAddr, ToSocketAddrs}, // 引入网络地址相关模块
    path::PathBuf,                    // 引入路径缓冲区模块
    sync::Arc,                        // 引入原子引用计数器模块
    time::Duration,                   // 引入时间持续模块
};

use gm_quic::{handy::*, *}; // 引入 gm_quic 模块及其辅助功能
use qevent::telemetry::handy::NullLogger; // 引入日志记录器模块
use tokio::{
    io::{self, AsyncWriteExt},
    net::TcpStream,
}; // 引入异步 I/O 模块
use tracing::{debug, info}; // 引入日志记录宏

#[tokio::main(flavor = "current_thread")] // 使用单线程 Tokio 运行时
async fn main() {
    // 初始化日志系统，使用环境变量过滤日志级别
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    // 从环境变量获取监听地址，若未提供则使用缺省值 "0.0.0.0:4433"
    // 支持多个地址以逗号分隔，并解析为 SocketAddr 类型的向量
    let listen_addresses = env::var("LISTEN_ADDRESSES")
        .unwrap_or_else(|_| "0.0.0.0:5100".to_string())
        .split(',')
        .map(|addr| addr.trim().parse::<SocketAddr>())
        .collect::<Result<Vec<SocketAddr>, _>>()
        .expect("Failed to parse LISTEN_ADDRESSES as SocketAddr");

    // 启动运行函数，传入解析后的监听地址
    if let Err(error) = run(listen_addresses.as_slice()).await {
        info!(?error, "server error"); // 记录服务器错误日志
        std::process::exit(1); // 退出程序
    }
}

// 运行 QUIC 服务器的异步函数
async fn run(addresses: impl ToSocketAddrs) -> io::Result<()> {
    // 创建一个空的日志记录器
    let qlogger: Arc<dyn qevent::telemetry::Log + Send + Sync> = Arc::new(NullLogger);

    // 配置 QUIC 服务器参数，设置最大空闲超时时间为 3600 秒
    let mut parameters = server_parameters();
    parameters.set_max_idle_timeout(Duration::from_secs(3600));

    // 从环境变量获取证书和私钥路径，若未提供则使用缺省值
    let cert_path = env::var("QUIC_CERT_PATH").unwrap_or_else(|_| "certs/wild.cer".to_string());
    let key_path = env::var("QUIC_KEY_PATH").unwrap_or_else(|_| "certs/wild.key".to_string());

    // 构建 QUIC 服务器，绑定监听地址并启动
    let server = QuicServer::builder()
        .without_client_cert_verifier() // 不验证客户端证书
        .with_single_cert(
            PathBuf::from(cert_path).as_path(), // 加载证书文件
            PathBuf::from(key_path).as_path(),  // 加载私钥文件
        )
        .with_parameters(parameters) // 应用服务器参数
        .with_qlog(qlogger) // 应用日志记录器
        .listen(addresses) // 绑定监听地址
        .unwrap();

    // 记录服务器监听的地址
    info!("listen on {:?}", server.addresses());

    // 启动代理转发服务
    serve_proxy(server).await
}

// 处理 QUIC 流的异步函数
async fn serve_proxy(server: Arc<QuicServer>) -> io::Result<()> {
    // 定义处理单个流的异步函数
    async fn handle_stream(mut reader: StreamReader, mut writer: StreamWriter) -> io::Result<()> {
        // 从环境变量获取目标地址，缺省为 "127.0.0.1:5100"
        let target_address =
            env::var("TARGET_ADDRESS").unwrap_or_else(|_| "127.0.0.1:5100".to_string());

        // 验证目标地址格式
        if target_address.parse::<SocketAddr>().is_err() {
            debug!("Invalid target address format: {}", target_address);
            writer.shutdown().await.ok();
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                "Invalid target address format",
            ));
        }

        // 建立到目标地址的 TCP 连接，添加超时
        let tcp_stream = match tokio::time::timeout(
            Duration::from_secs(2), // 2秒超时
            TcpStream::connect(&target_address),
        )
        .await
        {
            Ok(Ok(stream)) => stream,
            Ok(Err(e)) => {
                debug!("Failed to connect to {}: {}", target_address, e);
                writer.shutdown().await.ok();
                return Err(io::Error::new(
                    e.kind(),
                    format!("Failed to connect to target: {}", e),
                ));
            }
            Err(_) => {
                debug!("Connection to {} timed out", target_address);
                writer.shutdown().await.ok();
                return Err(io::Error::new(
                    io::ErrorKind::TimedOut,
                    "Connection to target timed out",
                ));
            }
        };

        // 将 TcpStream 拆分为独立的读写流
        let (mut tcp_reader, mut tcp_writer) = tcp_stream.into_split();

        // 创建双向转发任务
        let quic_to_tcp = async {
            let bytes = io::copy(&mut reader, &mut tcp_writer).await?;
            debug!("QUIC -> TCP: transferred {} bytes", bytes);
            tcp_writer.shutdown().await?;
            Ok::<u64, io::Error>(bytes)
        };

        let tcp_to_quic = async {
            let bytes = io::copy(&mut tcp_reader, &mut writer).await?;
            debug!("TCP -> QUIC: transferred {} bytes", bytes);
            writer.shutdown().await?;
            Ok::<u64, io::Error>(bytes)
        };

        // 并发执行双向转发任务
        match tokio::try_join!(quic_to_tcp, tcp_to_quic) {
            Ok((quic_bytes, tcp_bytes)) => {
                info!(
                    "Stream forwarding completed: QUIC->TCP: {} bytes, TCP->QUIC: {} bytes",
                    quic_bytes, tcp_bytes
                );
                Ok(())
            }
            Err(e) => {
                debug!("Stream forwarding error: {}", e);
                Err(io::Error::new(
                    io::ErrorKind::BrokenPipe,
                    format!("Stream forwarding failed: {}", e),
                ))
            }
        }
    }

    // 循环接受新的连接
    loop {
        let (connection, pathway) = server.accept().await?; // 接受新的连接
        info!(source = ?pathway.remote(), "accepted new connection"); // 记录新连接信息

        tokio::spawn(async move {
            // 循环接受双向流并处理
            while let Ok(Some((_sid, (reader, writer)))) = connection.accept_bi_stream().await {
                tokio::spawn(handle_stream(reader, writer)); // 启动异步任务处理流
            }
        });
    }
}
