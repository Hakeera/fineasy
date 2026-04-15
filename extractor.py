"""
extractor.py — extrai campos de boletos bancários em PDF.

Uso:
    python extractor.py <caminho_do_pdf>

Saída:
    JSON com os campos extraídos (stdout)
    Em caso de erro: JSON com campo "error"
"""

import sys
import json
import re
import pdfplumber


# ---------------------------------------------------------------------------
# Padrões regex
# ---------------------------------------------------------------------------

# CNPJ: XX.XXX.XXX/XXXX-XX
RE_CNPJ = re.compile(r"\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}")

# Valor monetário: R$ 1.234,56  ou  1.234,56
RE_VALOR = re.compile(r"R?\$?\s*(\d{1,3}(?:\.\d{3})*,\d{2})")

# Data: DD/MM/AAAA
RE_DATA = re.compile(r"\d{2}/\d{2}/\d{4}")

# Linha digitável / código de barras
# Formato: AAAAA.BBBBB CCCCC.CCCCCC DDDDD.DDDDDD E FFFFFFFFFFFFFFFFF
RE_LINHA_DIGITAVEL = re.compile(
    r"\d{5}\.\d{5}\s+\d{5}\.\d{6}\s+\d{5}\.\d{6}\s+\d\s+\d{14,15}"
)
# Código de barras numérico puro (47 ou 48 dígitos)
RE_COD_BARRAS = re.compile(r"\d{47,48}")


# ---------------------------------------------------------------------------
# Rótulos que costumam aparecer próximos ao vencimento
# ---------------------------------------------------------------------------
ROTULOS_VENCIMENTO = [
    "vencimento",
    "data de vencimento",
    "venc.",
    "venc ",
    "validade",
]

ROTULOS_VALOR = [
    "valor do documento",
    "valor cobrado",
    "valor a pagar",
    "valor",
    "(=) valor cobrado",
]


# ---------------------------------------------------------------------------
# Funções de extração
# ---------------------------------------------------------------------------

def extrair_cnpj(texto: str) -> str | None:
    """Retorna o primeiro CNPJ encontrado no texto."""
    match = RE_CNPJ.search(texto)
    return match.group(0) if match else None


def extrair_valor(texto: str, linhas: list[str]) -> str | None:
    """
    Tenta encontrar o valor buscando por rótulos conhecidos primeiro.
    Fallback: primeiro valor monetário do texto.
    """
    linhas_lower = [l.lower() for l in linhas]

    for i, linha in enumerate(linhas_lower):
        for rotulo in ROTULOS_VALOR:
            if rotulo in linha:
                # Procura o valor na mesma linha ou na próxima
                for candidata in linhas[i:i+3]:
                    match = RE_VALOR.search(candidata)
                    if match:
                        return match.group(1)

    # Fallback: primeiro valor monetário do texto inteiro
    match = RE_VALOR.search(texto)
    return match.group(1) if match else None


def extrair_vencimento(texto: str, linhas: list[str]) -> str | None:
    """
    Busca a data de vencimento procurando por rótulos conhecidos.
    Fallback: primeira data do texto.
    """
    linhas_lower = [l.lower() for l in linhas]

    for i, linha in enumerate(linhas_lower):
        for rotulo in ROTULOS_VENCIMENTO:
            if rotulo in linha:
                for candidata in linhas[i:i+3]:
                    match = RE_DATA.search(candidata)
                    if match:
                        return match.group(0)

    # Fallback: primeira data encontrada
    match = RE_DATA.search(texto)
    return match.group(0) if match else None


def extrair_codigo_barras(texto: str) -> str | None:
    """
    Tenta extrair a linha digitável formatada primeiro.
    Fallback: código de barras numérico puro (47-48 dígitos).
    Retorna sempre só os dígitos.
    """
    match = RE_LINHA_DIGITAVEL.search(texto)
    if match:
        # Remove espaços e pontos — retorna só dígitos
        return re.sub(r"[\s.]", "", match.group(0))

    match = RE_COD_BARRAS.search(texto)
    return match.group(0) if match else None


# ---------------------------------------------------------------------------
# Leitura do PDF
# ---------------------------------------------------------------------------

def extrair_texto_pdf(caminho: str) -> tuple[str, list[str]]:
    """
    Lê todas as páginas do PDF e retorna:
    - texto completo concatenado
    - lista de linhas (sem linhas vazias)
    """
    texto_completo = []

    with pdfplumber.open(caminho) as pdf:
        for pagina in pdf.pages:
            texto = pagina.extract_text()
            if texto:
                texto_completo.append(texto)

    texto = "\n".join(texto_completo)
    linhas = [l.strip() for l in texto.splitlines() if l.strip()]
    return texto, linhas


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Uso: python extractor.py <caminho_do_pdf>"}))
        sys.exit(1)

    caminho = sys.argv[1]

    try:
        texto, linhas = extrair_texto_pdf(caminho)
    except FileNotFoundError:
        print(json.dumps({"error": f"Arquivo não encontrado: {caminho}"}))
        sys.exit(1)
    except Exception as e:
        print(json.dumps({"error": f"Erro ao ler PDF: {str(e)}"}))
        sys.exit(1)

    resultado = {
        "cnpj":          extrair_cnpj(texto),
        "valor":         extrair_valor(texto, linhas),
        "vencimento":    extrair_vencimento(texto, linhas),
        "cod_barras":    extrair_codigo_barras(texto),
        "legivel":       True,   # PDF digital lido com sucesso
    }

    print(json.dumps(resultado, ensure_ascii=False))


if __name__ == "__main__":
    main()
