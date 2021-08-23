grammar gl;




expression
 : expression '[' expression ']'                              # IndexExpression
 | expression '.' identifierName                              # DotExpression
 | expression arguments                                       # FunctionCallExpression
 | '!' expression                                             # NotExpression
 | expression ( '==' | '!=') expression                       # EqualityExpression
 | expression '&&' expression                                 # LogicalAndExpression
 | expression '||' expression                                 # LogicalOrExpression
 | lvalue '=' expression                                      # AssignmentExpression
 | identifierName '->' expression                             # ClosureExpression
 | Identifier                                                 # IdentifierExpression
 | literal                                                    # LiteralExpression
 | '(' expression ')'                                         # ParenthesizedExpression
 ;


lvalue
 : identifierName
 ;


arguments
 : '(' argumentList? ')'
 ;

argumentList
 : expression ( ',' expression )*
 ;


literal
 : ( NullLiteral
   | BooleanLiteral
   | StringLiteral
   )
 | numericLiteral
 ;

numericLiteral
 : DecimalLiteral
 | HexIntegerLiteral
 ;

identifierName
 : Identifier
 ;


NullLiteral
 : 'null'
 ;

BooleanLiteral
 : 'true'
 | 'false'
 ;

DecimalLiteral
 : DecimalIntegerLiteral '.' DecimalDigit* ExponentPart?
 | '.' DecimalDigit+ ExponentPart?
 | DecimalIntegerLiteral ExponentPart?
 ;

HexIntegerLiteral
 : '0' [xX] HexDigit+
 ;


Identifier
 : IdentifierStart IdentifierPart*
 ;

StringLiteral
 : '"' StringCharacter* '"'
 | '\'' SingleStringCharacter* '\''
 ;

WhiteSpaces
 : [\t\u000B\u000C\u0020\u00A0]+ -> channel(HIDDEN)
 ;

fragment StringCharacter
 : ~["\\\r\n]
 | '\\' EscapeSequence
 ;

fragment SingleStringCharacter
 : ~['\\\r\n]
 | '\\' EscapeSequence
 ;

fragment EscapeSequence
 : CharacterEscapeSequence
 | HexEscapeSequence
 | UnicodeEscapeSequence
 ;

fragment CharacterEscapeSequence
 : SingleEscapeCharacter
 | NonEscapeCharacter
 ;

fragment HexEscapeSequence
 : 'x' HexDigit HexDigit
 ;

fragment UnicodeEscapeSequence
 : 'u' HexDigit HexDigit HexDigit HexDigit
 ;

fragment SingleEscapeCharacter
 : ['"\\bfnrtv]
 ;

fragment NonEscapeCharacter
 : ~['"\\bfnrtv0-9xu\r\n]
 ;

fragment EscapeCharacter
 : SingleEscapeCharacter
 | DecimalDigit
 | [xu]
 ;

fragment DecimalDigit
 : [0-9]
 ;

fragment HexDigit
 : [0-9a-fA-F]
 ;

fragment DecimalIntegerLiteral
 : '0'
 | [1-9] DecimalDigit*
 ;

fragment ExponentPart
 : [eE] [+-]? DecimalDigit+
 ;

fragment IdentifierStart
 : [\p{L}]
 | [$_]
 | '\\' UnicodeEscapeSequence
 ;

fragment IdentifierPart
 : IdentifierStart
 | [\p{Mn}]
 | [\p{Nd}]
 | [\p{Pc}]
 | ZWNJ
 | ZWJ
 ;

fragment ZWNJ
 : '\u200C'
 ;

fragment ZWJ
 : '\u200D'
 ;
