"""
extractor.py — extrai campos de boletos bancários e notas fiscais (DANFE) em PDF.

Uso CLI:
    python extractor.py <caminho_do_pdf>

Uso como módulo:
    from extractor import extrair_pdf
"""

import sys
import json
import re
import pdfplumber


# ---------------------------------------------------------------------------
# Regex
# ---------------------------------------------------------------------------

RE_CNPJ = re.compile(r"\d{2}\.?\d{3}\.?\d{3}/?\d{4}-?\d{2}")
RE_VALOR = re.compile(r"R?\$?\s*(\d{1,3}(?:\.\d{3})*,\d{2})")
RE_DATA  = re.compile(r"\d{2}/\d{2}/\d{4}")

RE_LINHA_DIGITAVEL = re.compile(
    r"\d{5}\.\d{5}\s+\d{5}\.\d{6}\s+\d{5}\.\d{6}\s+\d\s+\d{14,15}"
)
RE_COD_BARRAS = re.compile(r"\d{47,48}")
RE_CHAVE_NFE = re.compile(r"(?:\d{4}\s?){11}(?:\d{4})")


# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

MEUS_CNPJS = [
    "14518978000235",
    "14518978000154",
    "55409080000125",
    "57068174000103",
    "63659032000113",
]


# ---------------------------------------------------------------------------
# Utils
# ---------------------------------------------------------------------------

def normalizar_cnpj(cnpj: str) -> str:
    return re.sub(r"\D", "", cnpj)


def identificar_pagador_por_lista(cnpjs_encontrados, meus_cnpjs):
    meus_cnpjs_norm = {normalizar_cnpj(c) for c in meus_cnpjs}

    pagador = None
    emitente = None

    for cnpj in cnpjs_encontrados:
        cnpj_norm = normalizar_cnpj(cnpj)

        if cnpj_norm in meus_cnpjs_norm:
            pagador = cnpj
        else:
            if not emitente:
                emitente = cnpj

    return pagador, emitente


# ---------------------------------------------------------------------------
# Tipo doc
# ---------------------------------------------------------------------------

MARCADORES_DANFE = [
    "danfe",
    "documento auxiliar da nota fiscal eletrônica",
    "nota fiscal eletrônica",
    "nf-e",
    "chave de acesso",
]

MARCADORES_BOLETO = [
    "ficha de compensação",
    "boleto bancário",
    "boleto",
    "linha digitável",
    "beneficiário",
    "nosso número",
    "cedente",
]


def identificar_tipo(texto_lower):
    score_danfe  = sum(1 for m in MARCADORES_DANFE  if m in texto_lower)
    score_boleto = sum(1 for m in MARCADORES_BOLETO if m in texto_lower)

    if score_danfe >= 2:
        return "danfe"
    if score_boleto >= 2:
        return "boleto"
    if score_danfe > score_boleto:
        return "danfe"
    if score_boleto > score_danfe:
        return "boleto"
    return "desconhecido"


# ---------------------------------------------------------------------------
# Extrações
# ---------------------------------------------------------------------------

def extrair_texto_pdf(caminho):
    texto_completo = []

    with pdfplumber.open(caminho) as pdf:
        for pagina in pdf.pages:
            texto = pagina.extract_text()
            if texto:
                texto_completo.append(texto)

    texto = "\n".join(texto_completo)
    linhas = [l.strip() for l in texto.splitlines() if l.strip()]

    return texto, linhas


def extrair_valor(texto, linhas, tipo):
    rotulos = ["valor total", "valor a pagar"] if tipo == "danfe" else ["valor"]

    linhas_lower = [l.lower() for l in linhas]

    for i, linha in enumerate(linhas_lower):
        for rotulo in rotulos:
            if rotulo in linha:
                for candidata in linhas[i:i+3]:
                    match = RE_VALOR.search(candidata)
                    if match:
                        return match.group(1)

    match = RE_VALOR.search(texto)
    return match.group(1) if match else None


def extrair_vencimento(texto, linhas, tipo):
    rotulos = ["data de emissão"] if tipo == "danfe" else ["vencimento"]

    linhas_lower = [l.lower() for l in linhas]

    for i, linha in enumerate(linhas_lower):
        for rotulo in rotulos:
            if rotulo in linha:
                for candidata in linhas[i:i+3]:
                    match = RE_DATA.search(candidata)
                    if match:
                        return match.group(0)

    match = RE_DATA.search(texto)
    return match.group(0) if match else None


def extrair_codigo_barras(texto):
    match = RE_LINHA_DIGITAVEL.search(texto)
    if match:
        return re.sub(r"[\s.]", "", match.group(0))

    match = RE_COD_BARRAS.search(texto)
    return match.group(0) if match else None


def extrair_chave_nfe(texto):
    match = RE_CHAVE_NFE.search(texto)
    if match:
        return re.sub(r"\s", "", match.group(0))
    return None


# ---------------------------------------------------------------------------
# FUNÇÃO PRINCIPAL (IMPORTÁVEL)
# ---------------------------------------------------------------------------

def extrair_pdf(caminho_pdf: str) -> dict:
    try:
        texto, linhas = extrair_texto_pdf(caminho_pdf)
    except Exception as e:
        return {"error": str(e), "arquivo": caminho_pdf}

    tipo = identificar_tipo(texto.lower())

    todos_cnpjs = RE_CNPJ.findall(texto)
    cnpjs_unicos = list(dict.fromkeys(todos_cnpjs))

    cnpj_pagador_lista, cnpj_emitente_lista = identificar_pagador_por_lista(
        cnpjs_unicos,
        MEUS_CNPJS
    )

    resultado = {
        "tipo": tipo,
        "legivel": True,
        "arquivo": caminho_pdf,
        "cnpj_emitente": cnpj_emitente_lista,
        "cnpj_pagador": cnpj_pagador_lista,
        "valor": extrair_valor(texto, linhas, tipo),
        "vencimento": extrair_vencimento(texto, linhas, tipo),
        "cod_barras": extrair_codigo_barras(texto) if tipo == "boleto" else None,
        "chave_nfe": extrair_chave_nfe(texto) if tipo == "danfe" else None,
    }

    return resultado


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Uso: python extractor.py <pdf>"}))
        sys.exit(1)

    caminho = sys.argv[1]
    resultado = extrair_pdf(caminho)

    print(json.dumps(resultado, ensure_ascii=False))


if __name__ == "__main__":
    main()
