' ============================================================
' Módulo: Fineasy
' Descrição: Importa emails.csv e attachments.csv gerados pelo
'            script Go para planilhas do Excel.
' Como usar: Alt+F11 → Inserir → Módulo → colar este código
'            Depois: Alt+F8 → selecionar ImportarTudo → Executar
' ============================================================

Option Explicit

' ── Configuração ─────────────────────────────────────────────
' Altere este caminho para a pasta onde o executável Go gera os CSVs
Private Const DATA_FOLDER As String = "C:\fineasy\data\"

Private Const CSV_EMAILS      As String = "emails.csv"
Private Const CSV_ATTACHMENTS As String = "attachments.csv"

Private Const ABA_EMAILS      As String = "Emails"
Private Const ABA_ATTACHMENTS As String = "Attachments"

' ── Ponto de entrada principal ───────────────────────────────
Public Sub ImportarTudo()
    Application.ScreenUpdating = False
    Application.Calculation = xlCalculationManual

    ImportarCSV DATA_FOLDER & CSV_EMAILS, ABA_EMAILS
    ImportarCSV DATA_FOLDER & CSV_ATTACHMENTS, ABA_ATTACHMENTS
    FormatarAbas

    Application.Calculation = xlCalculationAutomatic
    Application.ScreenUpdating = True

    MsgBox "Importação concluída!" & vbCrLf & _
           "Emails importados: " & ContarLinhas(ABA_EMAILS) & vbCrLf & _
           "Attachments importados: " & ContarLinhas(ABA_ATTACHMENTS), _
           vbInformation, "Fineasy"
End Sub

' ── Importa um CSV para uma aba ──────────────────────────────
Private Sub ImportarCSV(caminhoCSV As String, nomeAba As String)
    Dim ws As Worksheet
    Dim fileNum As Integer
    Dim linha As String
    Dim campos() As String
    Dim linhaIdx As Long

    ' Verifica se o arquivo existe
    If Dir(caminhoCSV) = "" Then
        MsgBox "Arquivo não encontrado:" & vbCrLf & caminhoCSV, vbExclamation, "Fineasy"
        Exit Sub
    End If

    ' Cria ou limpa a aba
    ws = ObterOuCriarAba(nomeAba)
    ws.Cells.Clear

    fileNum = FreeFile
    Open caminhoCSV For Input As #fileNum

    linhaIdx = 1
    Do While Not EOF(fileNum)
        Line Input #fileNum, linha
        campos = ParseCSVLine(linha)

        Dim col As Integer
        For col = 0 To UBound(campos)
            ws.Cells(linhaIdx, col + 1).Value = campos(col)
        Next col

        linhaIdx = linhaIdx + 1
    Loop

    Close #fileNum
End Sub

' ── Aplica formatação visual nas abas ───────────────────────
Private Sub FormatarAbas()
    FormatarAba ABA_EMAILS
    FormatarAba ABA_ATTACHMENTS
End Sub

Private Sub FormatarAba(nomeAba As String)
    Dim ws As Worksheet
    On Error Resume Next
    Set ws = ThisWorkbook.Sheets(nomeAba)
    On Error GoTo 0
    If ws Is Nothing Then Exit Sub

    Dim ultimaLinha As Long
    Dim ultimaCol As Long
    ultimaLinha = ws.Cells(ws.Rows.Count, 1).End(xlUp).Row
    ultimaCol = ws.Cells(1, ws.Columns.Count).End(xlToLeft).Column

    If ultimaLinha < 1 Or ultimaCol < 1 Then Exit Sub

    ' Cabeçalho
    With ws.Range(ws.Cells(1, 1), ws.Cells(1, ultimaCol))
        .Font.Bold = True
        .Interior.Color = RGB(31, 78, 121)   ' azul escuro
        .Font.Color = RGB(255, 255, 255)      ' texto branco
    End With

    ' Linhas alternadas
    Dim i As Long
    For i = 2 To ultimaLinha
        If i Mod 2 = 0 Then
            ws.Range(ws.Cells(i, 1), ws.Cells(i, ultimaCol)).Interior.Color = RGB(217, 226, 243)
        Else
            ws.Range(ws.Cells(i, 1), ws.Cells(i, ultimaCol)).Interior.Color = RGB(255, 255, 255)
        End If
    Next i

    ' Auto-ajuste de colunas
    ws.Columns(1).AutoFit
    ws.Columns(2).AutoFit

    ' Coluna de assunto/filename mais larga
    If ultimaCol >= 3 Then ws.Columns(3).ColumnWidth = 45
    If ultimaCol >= 4 Then ws.Columns(4).ColumnWidth = 35
    If ultimaCol >= 5 Then ws.Columns(5).ColumnWidth = 22

    ' Congela cabeçalho
    ws.Activate
    ws.Rows(2).Select
    ActiveWindow.FreezePanes = True
    ws.Cells(1, 1).Select
End Sub

' ── Utilitários ──────────────────────────────────────────────

' Retorna a aba existente ou cria uma nova
Private Function ObterOuCriarAba(nome As String) As Worksheet
    Dim ws As Worksheet
    On Error Resume Next
    Set ws = ThisWorkbook.Sheets(nome)
    On Error GoTo 0

    If ws Is Nothing Then
        Set ws = ThisWorkbook.Sheets.Add(After:=ThisWorkbook.Sheets(ThisWorkbook.Sheets.Count))
        ws.Name = nome
    End If

    Set ObterOuCriarAba = ws
End Function

' Conta linhas de dados (exclui cabeçalho)
Private Function ContarLinhas(nomeAba As String) As Long
    Dim ws As Worksheet
    On Error Resume Next
    Set ws = ThisWorkbook.Sheets(nomeAba)
    On Error GoTo 0
    If ws Is Nothing Then
        ContarLinhas = 0
        Exit Function
    End If
    Dim ultima As Long
    ultima = ws.Cells(ws.Rows.Count, 1).End(xlUp).Row
    ContarLinhas = IIf(ultima > 1, ultima - 1, 0)
End Function

' Parser simples de linha CSV (trata campos entre aspas)
Private Function ParseCSVLine(linha As String) As String()
    Dim resultado() As String
    Dim campos As New Collection
    Dim i As Integer
    Dim c As String
    Dim campo As String
    Dim dentroAspas As Boolean

    dentroAspas = False
    campo = ""

    For i = 1 To Len(linha)
        c = Mid(linha, i, 1)
        If c = Chr(34) Then                         ' aspas duplas
            If dentroAspas And Mid(linha, i + 1, 1) = Chr(34) Then
                campo = campo & Chr(34)             ' aspas escapadas ""
                i = i + 1
            Else
                dentroAspas = Not dentroAspas
            End If
        ElseIf c = "," And Not dentroAspas Then
            campos.Add campo
            campo = ""
        Else
            campo = campo & c
        End If
    Next i
    campos.Add campo                                ' último campo

    ReDim resultado(campos.Count - 1)
    Dim idx As Integer
    idx = 0
    Dim item As Variant
    For Each item In campos
        resultado(idx) = CStr(item)
        idx = idx + 1
    Next item

    ParseCSVLine = resultado
End Function
