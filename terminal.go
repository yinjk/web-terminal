package main

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "sync"

    "gopkg.in/igm/sockjs-go.v2/sockjs"
    "golang.org/x/crypto/ssh"
)

const END_OF_TRANSMISSION = "\u0004"

// TerminalSize represents the width and height of a terminal.
type TerminalSize struct {
    Width  uint16
    Height uint16
}

// TerminalSizeQueue is capable of returning terminal resize events as they occur.
type TerminalSizeQueue interface {
    // Next returns the new terminal size after the terminal has been resized. It returns nil when
    // monitoring has been stopped.
    Next() *TerminalSize
}


// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
    io.Reader
    io.Writer
    TerminalSizeQueue
}

// TerminalSession implements PtyHandler (using a SockJS connection)
type TerminalSession struct {
    id            string
    bound         chan error
    sockJSSession sockjs.Session
    sizeChan      chan TerminalSize
}

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
    Op, Data, SessionID string
    Rows, Cols          uint16
}

// TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Next() *TerminalSize {
    select {
    case size := <-t.sizeChan:
        return &size
    }
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Read(p []byte) (int, error) {
    m, err := t.sockJSSession.Recv()
    if err != nil {
        // Send terminated signal to process to avoid resource leak
        return copy(p, END_OF_TRANSMISSION), err
    }

    var msg TerminalMessage
    if err := json.Unmarshal([]byte(m), &msg); err != nil {
        return copy(p, END_OF_TRANSMISSION), err
    }

    switch msg.Op {
    case "stdin":
        return copy(p, msg.Data), nil
    case "resize":
        t.sizeChan <- TerminalSize{Width: msg.Cols, Height: msg.Rows}
        return 0, nil
    default:
        return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
    }
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t TerminalSession) Write(p []byte) (int, error) {
    msg, err := json.Marshal(TerminalMessage{
        Op:   "stdout",
        Data: string(p),
    })
    if err != nil {
        return 0, err
    }

    if err = t.sockJSSession.Send(string(msg)); err != nil {
        return 0, err
    }
    return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t TerminalSession) Toast(p string) error {
    msg, err := json.Marshal(TerminalMessage{
        Op:   "toast",
        Data: p,
    })
    if err != nil {
        return err
    }

    if err = t.sockJSSession.Send(string(msg)); err != nil {
        return err
    }
    return nil
}

// SessionMap stores a map of all TerminalSession objects and a lock to avoid concurrent conflict
type SessionMap struct {
    Sessions map[string]TerminalSession
    Lock     sync.RWMutex
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) TerminalSession {
    sm.Lock.RLock()
    defer sm.Lock.RUnlock()
    return sm.Sessions[sessionId]
}

// Set store a TerminalSession to SessionMap
func (sm *SessionMap) Set(sessionId string, session TerminalSession) {
    sm.Lock.Lock()
    defer sm.Lock.Unlock()
    sm.Sessions[sessionId] = session
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
    sm.Lock.Lock()
    defer sm.Lock.Unlock()
    sm.Sessions[sessionId].sockJSSession.Close(status, reason)
    delete(sm.Sessions, sessionId)
}

var terminalSessions = SessionMap{Sessions: make(map[string]TerminalSession)}

// handleTerminalSession is Called by net/http for any new /api/sockjs connections
func handleTerminalSession(session sockjs.Session) {
    var (
        buf             string
        err             error
        msg             TerminalMessage
        terminalSession TerminalSession
    )

    if buf, err = session.Recv(); err != nil {
        log.Printf("handleTerminalSession: can't Recv: %v", err)
        return
    }

    if err = json.Unmarshal([]byte(buf), &msg); err != nil {
        log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
        return
    }

    if msg.Op != "bind" {
        log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
        return
    }

    if terminalSession = terminalSessions.Get(msg.SessionID); terminalSession.id == "" {
        log.Printf("handleTerminalSession: can't find session '%s'", msg.SessionID)
        return
    }

    terminalSession.sockJSSession = session
    terminalSessions.Set(msg.SessionID, terminalSession)
    terminalSession.bound <- nil
}

// CreateAttachHandler is called from main for /api/sockjs
func CreateAttachHandler(path string) http.Handler {
    return sockjs.NewHandler(path, sockjs.DefaultOptions, handleTerminalSession)
}


// genTerminalSessionId generates a random session ID string. The format is not really interesting.
// This ID is used to identify the session when the client opens the SockJS connection.
// Not the same as the SockJS session id! We can't use that as that is generated
// on the client side and we don't have it yet at this point.
func genTerminalSessionId() (string, error) {
    bytes := make([]byte, 16)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    id := make([]byte, hex.EncodedLen(len(bytes)))
    hex.Encode(id, bytes)
    return string(id), nil
}

// isValidShell checks if the shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
    for _, validShell := range validShells {
        if validShell == shell {
            return true
        }
    }
    return false
}

/**
 * 等待node终端连接
 * @param :
 * @return:
 * @author: 尹健康
 * @time  : 2019/3/21 14:56
 */
func WaitForNodeTerminal(ip, username string, port int, password, sessionId string) {
    select {
    case <-terminalSessions.Get(sessionId).bound:
        close(terminalSessions.Get(sessionId).bound)

        var err error

        err = startNodeProcess(username, ip, port, password, terminalSessions.Get(sessionId))

        if err != nil {
            terminalSessions.Close(sessionId, 2, err.Error())
            return
        }

        terminalSessions.Close(sessionId, 1, "Process exited")
    }
}

/**
 * 开始node shell连接进程
 * @param :
 * @return: 
 * @author: 尹健康
 * @time  : 2019/3/21 14:57
 */
func startNodeProcess(username, ip string, port int, password string, ptyHandler PtyHandler) error {
    session, err := sshConnect(username, password, ip, port)
    if err != nil {
        panic(err) //TODO panic处理一下
        return err
    }
    defer session.Close()

    // Set up terminal modes
    modes := ssh.TerminalModes{
        ssh.ECHO:          1,     // disable echoing
        ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
        ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
    }
    // Request pseudo terminal
    //其实高没什么影响，宽设置的大一点，不然超过限制的字符会自动跳到行首
    if err := session.RequestPty("xterm", 20, 400, modes); err != nil {
        return err
    }
    session.Stdout = ptyHandler
    session.Stderr = ptyHandler
    session.Stdin = ptyHandler
    if err := session.Shell(); nil != err {
        panic(err)
        return err
    }
    if err := session.Wait(); nil != err {
        panic(err)
        return err
    }
    return nil
}